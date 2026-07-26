// Package render turns an ecs.World into draw calls. It runs one pass per layer
// (Background -> Entities -> UI), and within each layer sorts by Order.Z so
// draw order is deterministic despite ECS archetype iteration order. The
// Background and Entities layers are offset by the World's Camera resource (if
// any); the UI layer is always drawn in screen space.
package render

import (
	"math"
	"sort"

	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/types"
)

// Drawer is the minimal drawing surface the renderer needs. canvas.Canvas
// satisfies it; tests supply a mock.
type Drawer interface {
	DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error
}

type item struct {
	texturePath string
	position    types.Vector2
	size        types.Vector2
	uv          types.UVRect
	z           int
}

// Renderer holds one filter per layer, built once against a World. Reuse it for
// the lifetime of that World; rebuild when the active State (and its World)
// changes.
type Renderer struct {
	world *ecs.World
	bg    *ecs.Filter3[components.Position, components.Sprite, components.Order]
	ent   *ecs.Filter3[components.Position, components.Sprite, components.Order]
	ui    *ecs.Filter3[components.Position, components.Sprite, components.Order]
	buf   []item
}

// NewRenderer builds a Renderer for world w.
func NewRenderer(w *ecs.World) *Renderer {
	return &Renderer{
		world: w,
		bg:    ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerBackground]()),
		ent:   ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerEntities]()),
		ui:    ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerUI]()),
	}
}

// Draw renders all visible entities to d in layer then Order.Z order.
func (r *Renderer) Draw(d Drawer) {
	cam := ecs.GetResource[components.Camera](r.world)
	r.pass(r.bg, d, cam)
	r.pass(r.ent, d, cam)
	r.pass(r.ui, d, nil) // UI is drawn in screen space; the camera never affects it.
}

func (r *Renderer) pass(f *ecs.Filter3[components.Position, components.Sprite, components.Order], d Drawer, cam *components.Camera) {
	r.buf = r.buf[:0]

	zoom := 1.0
	var camX, camY float64
	if cam != nil {
		camX, camY = cam.X, cam.Y
		if cam.Zoom > 0 {
			zoom = cam.Zoom
		}
	}

	f.Each(func(_ ecs.Entity, p *components.Position, sp *components.Sprite, o *components.Order) {
		if !sp.Visible {
			return
		}
		// Snap each edge independently so adjacent tiles share a pixel boundary
		// under fractional zoom (avoids 1px gaps / overlaps from rounding size
		// and position separately).
		x0 := math.Round((p.X - camX) * zoom)
		y0 := math.Round((p.Y - camY) * zoom)
		x1 := math.Round((p.X + sp.Size.X - camX) * zoom)
		y1 := math.Round((p.Y + sp.Size.Y - camY) * zoom)
		r.buf = append(r.buf, item{
			texturePath: sp.TexturePath,
			position:    types.Vector2{X: x0, Y: y0},
			size:        types.Vector2{X: x1 - x0, Y: y1 - y0},
			uv:          sp.UV(),
			z:           o.Z,
		})
	})

	sort.SliceStable(r.buf, func(i, j int) bool { return r.buf[i].z < r.buf[j].z })

	for i := range r.buf {
		it := &r.buf[i]
		// Texture may not be loaded yet; ignore per-draw errors like the old path.
		_ = d.DrawTexturedRect(it.texturePath, it.position, it.size, it.uv)
	}
}

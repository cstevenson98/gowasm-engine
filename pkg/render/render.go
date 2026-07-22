// Package render turns an ecs.World into draw calls. It runs one pass per layer
// (Background -> Entities -> UI), and within each layer sorts by Order.Z so
// draw order is deterministic despite ECS archetype iteration order.
package render

import (
	"sort"

	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
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
	bg  *ecs.Filter3[components.Position, components.Sprite, components.Order]
	ent *ecs.Filter3[components.Position, components.Sprite, components.Order]
	ui  *ecs.Filter3[components.Position, components.Sprite, components.Order]
	buf []item
}

// NewRenderer builds a Renderer for world w.
func NewRenderer(w *ecs.World) *Renderer {
	return &Renderer{
		bg:  ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerBackground]()),
		ent: ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerEntities]()),
		ui:  ecs.NewFilter3[components.Position, components.Sprite, components.Order](w).With(ecs.C[components.LayerUI]()),
	}
}

// Draw renders all visible entities to d in layer then Order.Z order.
func (r *Renderer) Draw(d Drawer) {
	r.pass(r.bg, d)
	r.pass(r.ent, d)
	r.pass(r.ui, d)
}

func (r *Renderer) pass(f *ecs.Filter3[components.Position, components.Sprite, components.Order], d Drawer) {
	r.buf = r.buf[:0]
	f.Each(func(_ ecs.Entity, p *components.Position, sp *components.Sprite, o *components.Order) {
		if !sp.Visible {
			return
		}
		r.buf = append(r.buf, item{
			texturePath: sp.TexturePath,
			position:    types.Vector2{X: p.X, Y: p.Y},
			size:        sp.Size,
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

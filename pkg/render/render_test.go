package render

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

type mockDrawer struct{ calls []string }

func (m *mockDrawer) DrawTexturedRect(tp string, _ types.Vector2, _ types.Vector2, _ types.UVRect) error {
	m.calls = append(m.calls, tp)
	return nil
}

func spriteNamed(name string, visible bool) *components.Sprite {
	return &components.Sprite{TexturePath: name, Columns: 1, Rows: 1, Visible: visible}
}

func TestRendererLayerAndOrder(t *testing.T) {
	w := ecs.NewWorld()

	bg := ecs.NewMap4[components.Position, components.Sprite, components.LayerBackground, components.Order](w)
	en := ecs.NewMap4[components.Position, components.Sprite, components.LayerEntities, components.Order](w)
	ui := ecs.NewMap4[components.Position, components.Sprite, components.LayerUI, components.Order](w)

	// Spawn out of visual order to prove layer + Z sorting, not insertion order.
	ui.NewEntity(&components.Position{}, spriteNamed("ui", true), &components.LayerUI{}, &components.Order{Z: 0})
	en.NewEntity(&components.Position{}, spriteNamed("e_hi", true), &components.LayerEntities{}, &components.Order{Z: 5})
	bg.NewEntity(&components.Position{}, spriteNamed("bg", true), &components.LayerBackground{}, &components.Order{Z: 0})
	en.NewEntity(&components.Position{}, spriteNamed("e_lo", true), &components.LayerEntities{}, &components.Order{Z: 1})

	d := &mockDrawer{}
	NewRenderer(w).Draw(d)

	want := []string{"bg", "e_lo", "e_hi", "ui"}
	if len(d.calls) != len(want) {
		t.Fatalf("draw calls = %v, want %v", d.calls, want)
	}
	for i := range want {
		if d.calls[i] != want[i] {
			t.Fatalf("draw order = %v, want %v", d.calls, want)
		}
	}
}

func TestRendererSkipsInvisible(t *testing.T) {
	w := ecs.NewWorld()
	en := ecs.NewMap4[components.Position, components.Sprite, components.LayerEntities, components.Order](w)
	en.NewEntity(&components.Position{}, spriteNamed("visible", true), &components.LayerEntities{}, &components.Order{})
	en.NewEntity(&components.Position{}, spriteNamed("hidden", false), &components.LayerEntities{}, &components.Order{})

	d := &mockDrawer{}
	NewRenderer(w).Draw(d)

	if len(d.calls) != 1 || d.calls[0] != "visible" {
		t.Fatalf("draw calls = %v, want [visible]", d.calls)
	}
}

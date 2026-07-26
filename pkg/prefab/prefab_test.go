package prefab

import (
	"testing"

	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/systems"
	"github.com/cstevenson98/milo/pkg/types"
)

func TestNewBackgroundComponents(t *testing.T) {
	w := ecs.NewWorld()
	e := NewBackground(w, types.Vector2{X: 1, Y: 2}, types.Vector2{X: 320, Y: 240}, "bg.png")

	sm := ecs.NewMap1[components.Sprite](w)
	if !sm.Has(e) {
		t.Fatal("background should have a Sprite")
	}
	sp := sm.Get(e)
	if sp.TexturePath != "bg.png" || sp.Columns != 1 || sp.Rows != 1 || !sp.Visible {
		t.Fatalf("background sprite = %+v", *sp)
	}
	// Background must be static: no Velocity, no Animation.
	if ecs.NewMap1[components.Velocity](w).Has(e) {
		t.Fatal("background should not have Velocity")
	}
	if ecs.NewMap1[components.Animation](w).Has(e) {
		t.Fatal("background should not have Animation")
	}
	if !ecs.NewMap1[components.LayerBackground](w).Has(e) {
		t.Fatal("background should be on BACKGROUND layer")
	}
}

func TestNewLlamaMovesAndAnimates(t *testing.T) {
	w := ecs.NewWorld()
	ecs.SetResource(w, &components.ScreenBounds{W: 320, H: 240})

	e := NewLlama(w, types.Vector2{X: 10, Y: 20}, types.Vector2{X: 16, Y: 16}, 50, 0.1)

	move := systems.NewMovement(w)
	anim := systems.NewAnimation(w)

	// One second of movement at speed 50 -> X advances by 50.
	move.Update(w, 1.0)

	pm := ecs.NewMap1[components.Position](w)
	if p := pm.Get(e); p.X != 60 {
		t.Fatalf("llama X after 1s = %v, want 60", p.X)
	}

	// frameTime = 0.1 + (50/100)*0.1 = 0.15s. Advance one frame (not a full cycle).
	anim.Update(w, 0.2)
	sm := ecs.NewMap1[components.Sprite](w)
	if sp := sm.Get(e); sp.Frame != 1 {
		t.Fatalf("llama frame = %d, want 1", sp.Frame)
	}
}

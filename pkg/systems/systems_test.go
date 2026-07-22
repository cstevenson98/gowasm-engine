package systems

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

func TestMovementIntegrates(t *testing.T) {
	w := ecs.NewWorld()
	m := ecs.NewMap2[components.Position, components.Velocity](w)
	e := m.NewEntity(&components.Position{X: 0, Y: 0}, &components.Velocity{DX: 100, DY: -50})

	sys := NewMovement(w)
	sys.Update(w, 0.5)

	p, _ := m.Get(e)
	if p.X != 50 || p.Y != -25 {
		t.Fatalf("position = %+v, want {50 -25}", *p)
	}
}

func TestMovementWrapsRight(t *testing.T) {
	w := ecs.NewWorld()
	ecs.SetResource(w, &components.ScreenBounds{W: 320, H: 240})

	m := ecs.NewMap3[components.Position, components.Velocity, components.Wrap](w)
	e := m.NewEntity(
		&components.Position{X: 319, Y: 10},
		&components.Velocity{DX: 100, DY: 0},
		&components.Wrap{SpriteW: 16, SpriteH: 16},
	)

	sys := NewMovement(w)
	sys.Update(w, 1.0) // moves to X=419, beyond W=320 -> wraps to -16

	p, _, _ := m.Get(e)
	if p.X != -16 {
		t.Fatalf("wrapped X = %v, want -16", p.X)
	}
}

func TestMovementNoWrapWithoutResource(t *testing.T) {
	w := ecs.NewWorld() // no ScreenBounds resource
	m := ecs.NewMap3[components.Position, components.Velocity, components.Wrap](w)
	e := m.NewEntity(
		&components.Position{X: 319},
		&components.Velocity{DX: 100},
		&components.Wrap{SpriteW: 16, SpriteH: 16},
	)

	NewMovement(w).Update(w, 1.0) // integrates to 419, no wrap (no bounds)

	p, _, _ := m.Get(e)
	if p.X != 419 {
		t.Fatalf("X = %v, want 419 (no wrap without bounds)", p.X)
	}
}

func TestAnimationCycles(t *testing.T) {
	w := ecs.NewWorld()
	m := ecs.NewMap2[components.Sprite, components.Animation](w)
	e := m.NewEntity(
		&components.Sprite{Columns: 2, Rows: 3}, // 6 frames
		&components.Animation{FrameTime: 0.1},
	)

	sys := NewAnimation(w)

	sys.Update(w, 0.05) // not enough for a frame
	sp, _ := m.Get(e)
	if sp.Frame != 0 {
		t.Fatalf("frame after 0.05s = %d, want 0", sp.Frame)
	}

	sys.Update(w, 0.07) // total 0.12 -> one frame advance
	sp, _ = m.Get(e)
	if sp.Frame != 1 {
		t.Fatalf("frame after 0.12s = %d, want 1", sp.Frame)
	}

	sys.Update(w, 0.25) // +2.5 frames -> +2 -> frame 3
	sp, _ = m.Get(e)
	if sp.Frame != 3 {
		t.Fatalf("frame = %d, want 3", sp.Frame)
	}
}

func TestAnimationStaticSpriteIgnored(t *testing.T) {
	w := ecs.NewWorld()
	m := ecs.NewMap2[components.Sprite, components.Animation](w)
	e := m.NewEntity(
		&components.Sprite{Columns: 1, Rows: 1}, // single frame
		&components.Animation{FrameTime: 0.1},
	)
	NewAnimation(w).Update(w, 1.0)
	sp, _ := m.Get(e)
	if sp.Frame != 0 {
		t.Fatalf("static sprite frame = %d, want 0", sp.Frame)
	}
}

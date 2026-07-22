package entities

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

// setInput sets the Input resource to the given state.
func setInput(w *ecs.World, st types.InputState) {
	ecs.SetResource(w, &components.Input{State: st})
}

func TestSpawnPlayerComponents(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnPlayer(w, types.Vector2{X: 10, Y: 20}, types.Vector2{X: 64, Y: 64}, 150,
		Stats{Level: 1, HP: 100, MaxHP: 100})

	if pc := ecs.NewMap1[PlayerControl](w).Get(e); pc.Speed != 150 {
		t.Fatalf("PlayerControl.Speed = %v, want 150", pc.Speed)
	}
	if !ecs.NewMap1[components.Velocity](w).Has(e) {
		t.Fatal("player should have Velocity")
	}
	if s := ecs.NewMap1[Stats](w).Get(e); s == nil || s.HP != 100 {
		t.Fatalf("player Stats = %+v, want HP 100", s)
	}
}

func TestPlayerInputSingleDirection(t *testing.T) {
	cases := []struct {
		name  string
		st    types.InputState
		wantX float64
		wantY float64
	}{
		{"up", types.InputState{MoveUp: true}, 0, -100},
		{"down", types.InputState{MoveDown: true}, 0, 100},
		{"left", types.InputState{MoveLeft: true}, -100, 0},
		{"right", types.InputState{MoveRight: true}, 100, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := ecs.NewWorld()
			e := SpawnPlayer(w, types.Vector2{}, types.Vector2{X: 64, Y: 64}, 100, Stats{})
			setInput(w, c.st)

			NewPlayerInputSystem(w).Update(w, 0)

			v := ecs.NewMap1[components.Velocity](w).Get(e)
			if !floatEquals(v.DX, c.wantX, 0.001) || !floatEquals(v.DY, c.wantY, 0.001) {
				t.Fatalf("velocity = (%v, %v), want (%v, %v)", v.DX, v.DY, c.wantX, c.wantY)
			}
		})
	}
}

func TestPlayerInputDiagonalNormalized(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnPlayer(w, types.Vector2{}, types.Vector2{X: 64, Y: 64}, 100, Stats{})
	setInput(w, types.InputState{MoveUp: true, MoveRight: true})

	NewPlayerInputSystem(w).Update(w, 0)

	v := ecs.NewMap1[components.Velocity](w).Get(e)
	want := 100.0 * 0.7071
	if !floatEquals(v.DX, want, 0.01) || !floatEquals(v.DY, -want, 0.01) {
		t.Fatalf("diagonal velocity = (%v, %v), want (%v, %v)", v.DX, v.DY, want, -want)
	}
	total := v.DX*v.DX + v.DY*v.DY
	if !floatEquals(total, 100.0*100.0, 1.0) {
		t.Fatalf("diagonal speed not preserved: %v", total)
	}
}

func TestPlayerInputOppositeCancels(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnPlayer(w, types.Vector2{}, types.Vector2{X: 64, Y: 64}, 100, Stats{})
	setInput(w, types.InputState{MoveUp: true, MoveDown: true, MoveLeft: true, MoveRight: true})

	NewPlayerInputSystem(w).Update(w, 0)

	v := ecs.NewMap1[components.Velocity](w).Get(e)
	if v.DX != 0 || v.DY != 0 {
		t.Fatalf("opposite directions should cancel, got (%v, %v)", v.DX, v.DY)
	}
}

func TestParticipantBattleEntity(t *testing.T) {
	var _ interface {
		GetID() string
	} = NewParticipant("Player", 100, 100, types.Vector2{X: 1, Y: 2}, true)

	p := NewParticipant("Player", 80, 100, types.Vector2{X: 5, Y: 6}, true)
	if p.GetID() != "Player" {
		t.Fatalf("id = %q", p.GetID())
	}
	if p.GetStats().HP != 80 {
		t.Fatalf("HP = %d, want 80", p.GetStats().HP)
	}
	if p.IsReady() {
		t.Fatal("new participant should not be ready")
	}
	p.ChargeTimer(2.0) // charge past full
	if !p.IsReady() {
		t.Fatal("participant should be ready after charging")
	}
	p.ResetTimer()
	if p.IsReady() {
		t.Fatal("participant should not be ready after reset")
	}
	if pos := p.GetPosition(); pos.X != 5 || pos.Y != 6 {
		t.Fatalf("position = %+v, want (5,6)", pos)
	}
}

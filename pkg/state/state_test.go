package state

import (
	"testing"

	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
)

// testState is a minimal concrete State used to exercise BaseState.
type testState struct {
	*BaseState
	spawned ecs.Entity
}

func newTestState() *testState { return &testState{BaseState: NewBaseState("Test")} }

func (s *testState) Enter(deps Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}
	m := ecs.NewMap2[components.Position, components.Velocity](s.World())
	s.spawned = m.NewEntity(&components.Position{}, &components.Velocity{DX: 10})
	s.Schedule().Add(&moveSystem{f: ecs.NewFilter2[components.Position, components.Velocity](s.World())})
	return nil
}

type moveSystem struct {
	f *ecs.Filter2[components.Position, components.Velocity]
}

func (m *moveSystem) Update(w *ecs.World, dt float64) {
	m.f.Each(func(_ ecs.Entity, p *components.Position, v *components.Velocity) {
		p.X += v.DX * dt
	})
}

func TestBaseStateLifecycle(t *testing.T) {
	s := newTestState()

	if s.Name() != "Test" {
		t.Fatalf("Name() = %q", s.Name())
	}

	deps := Deps{ScreenWidth: 800, ScreenHeight: 600}
	if err := s.Enter(deps); err != nil {
		t.Fatalf("Enter: %v", err)
	}

	// ScreenBounds resource seeded from deps.
	if b := ecs.GetResource[components.ScreenBounds](s.World()); b == nil || b.W != 800 || b.H != 600 {
		t.Fatalf("ScreenBounds = %+v, want 800x600", b)
	}

	// Deps stored.
	if s.Deps().ScreenWidth != 800 {
		t.Fatalf("Deps().ScreenWidth = %v", s.Deps().ScreenWidth)
	}

	// Update runs the schedule.
	s.Update(1.0)
	m := ecs.NewMap2[components.Position, components.Velocity](s.World())
	p, _ := m.Get(s.spawned)
	if p.X != 10 {
		t.Fatalf("after Update, Position.X = %v, want 10", p.X)
	}

	// Exit clears the world: no entities remain.
	s.Exit()
	if n := ecs.NewFilter1[components.Position](s.World()).Count(); n != 0 {
		t.Fatalf("after Exit, %d entities remain, want 0", n)
	}
}

// Ensure testState satisfies the State interface.
var _ State = (*testState)(nil)

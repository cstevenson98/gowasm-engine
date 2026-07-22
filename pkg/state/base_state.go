package state

import (
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// BaseState provides the boilerplate for a State: it owns the World and an
// ordered system Schedule, stores injected Deps, and seeds common resources.
// Embed it in a concrete state and override Enter to build entities and add
// systems.
//
//	type MenuState struct{ *state.BaseState }
//
//	func NewMenuState() *MenuState { return &MenuState{state.NewBaseState("Menu")} }
//
//	func (s *MenuState) Enter(deps state.Deps) error {
//	    if err := s.BaseState.Enter(deps); err != nil { return err }
//	    s.Schedule().Add(systems.NewMovement(s.World()))
//	    // ... build entities ...
//	    return nil
//	}
type BaseState struct {
	name  string
	world *ecs.World
	sched *ecs.Schedule
	deps  Deps
}

// NewBaseState creates a BaseState with a fresh World and empty Schedule.
func NewBaseState(name string) *BaseState {
	return &BaseState{
		name:  name,
		world: ecs.NewWorld(),
		sched: ecs.NewSchedule(),
	}
}

// Name returns the state identifier.
func (b *BaseState) Name() string { return b.name }

// World returns the state's ECS world.
func (b *BaseState) World() *ecs.World { return b.world }

// Schedule returns the state's system schedule, for registering systems in Enter.
func (b *BaseState) Schedule() *ecs.Schedule { return b.sched }

// Deps returns the engine services injected at Enter.
func (b *BaseState) Deps() Deps { return b.deps }

// Enter stores the injected deps and seeds the ScreenBounds resource. Concrete
// states should call BaseState.Enter first, then build entities and add systems.
func (b *BaseState) Enter(deps Deps) error {
	b.deps = deps
	ecs.SetResource(b.world, &components.ScreenBounds{W: deps.ScreenWidth, H: deps.ScreenHeight})
	return nil
}

// Update runs the schedule against the world. Override only if a state needs
// custom ordering beyond the schedule.
func (b *BaseState) Update(dt float64) {
	b.sched.Run(b.world, dt)
}

// Exit resets the world, dropping all entities and resources. The World object
// is retained so a re-entered state starts clean.
func (b *BaseState) Exit() {
	b.world.Reset()
}

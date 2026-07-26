package state

import (
	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/debug"
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/logger"
	"github.com/cstevenson98/milo/pkg/types"
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

// Enter stores the injected deps and seeds the ScreenBounds, Input, and Camera
// resources. Concrete states should call BaseState.Enter first, then build
// entities and add systems. The seeded Camera is an identity camera (0,0, zoom
// 1), so states that never add a CameraFollow system (or move the camera
// themselves) render exactly as if there were no camera at all.
func (b *BaseState) Enter(deps Deps) error {
	b.deps = deps
	ecs.SetResource(b.world, &components.ScreenBounds{W: deps.ScreenWidth, H: deps.ScreenHeight})
	ecs.SetResource(b.world, &components.Input{})
	ecs.SetResource(b.world, &components.Camera{Zoom: 1})
	return nil
}

// Update runs the schedule against the world, then drives the shared debug
// console (toggle on key 3, message aging). Override to add custom logic, but
// call BaseState.Update so the console keeps working everywhere.
func (b *BaseState) Update(dt float64) {
	b.sched.Run(b.world, dt)
	b.updateDebugConsole(dt)
}

func (b *BaseState) updateDebugConsole(dt float64) {
	if !b.deps.Debug.Enabled {
		return
	}
	in := b.Input()
	if in.Key3Pressed && !in.Key3PressedLastFrame {
		debug.Console.ToggleVisibility()
		logger.Logger.Debugf("Debug console toggled via key 3")
	}
	debug.Console.Update(dt)
}

// Exit resets the world, dropping all entities and resources. The World object
// is retained so a re-entered state starts clean.
func (b *BaseState) Exit() {
	b.world.Reset()
}

// DrawOverlays renders the debug console. Concrete states that show menus/HUD
// override this and call BaseState.DrawOverlays() to keep the console.
func (b *BaseState) DrawOverlays() error {
	return debug.Console.Render(b.UI())
}

// ===== Dependency accessors =====

// Input returns the latest input snapshot, or a zero value if no capturer is set.
func (b *BaseState) Input() types.InputState {
	if b.deps.Input != nil {
		return b.deps.Input.GetInputState()
	}
	return types.InputState{}
}

// UI returns the engine UI manager for overlay drawing (never nil).
func (b *BaseState) UI() types.UIManager {
	if b.deps.UI != nil {
		return b.deps.UI
	}
	return types.NopUI
}

// RequestState asks the engine to switch to another game state next frame.
func (b *BaseState) RequestState(s types.GameState) error {
	if b.deps.RequestState != nil {
		return b.deps.RequestState(s)
	}
	return nil
}

// GameStateProvider returns the game-defined global state provider, or nil.
func (b *BaseState) GameStateProvider() interface{} { return b.deps.GameState }

// ScreenWidth returns the virtual screen width.
func (b *BaseState) ScreenWidth() float64 { return b.deps.ScreenWidth }

// ScreenHeight returns the virtual screen height.
func (b *BaseState) ScreenHeight() float64 { return b.deps.ScreenHeight }

// DefaultFrameTime returns the engine's fallback seconds-per-frame for
// animated sprites (config.Settings.Animation.DefaultFrameTime).
func (b *BaseState) DefaultFrameTime() float64 { return b.deps.DefaultFrameTime }

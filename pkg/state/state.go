// Package state defines the top-level game-state abstraction that the engine
// swaps between (menu, gameplay, battle, ...). It is the ECS-era replacement
// for the old scene package: a State owns one ecs.World and a schedule of
// systems, built up in Enter and torn down in Exit.
//
// "State" is deliberate nomenclature: these are game states / screens keyed by
// types.GameState, not scene-graph nodes.
package state

import (
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/types"
)

// DebugConfig is the subset of the engine's debug settings a State needs. The
// engine populates it from its own config.Settings when entering a state, so
// pkg/state has no dependency on the engine's config package.
type DebugConfig struct {
	Enabled bool
}

// Deps are the engine services handed to a State when it becomes active. They
// are stored on the World as resources by BaseState.Enter so systems can reach
// them without bespoke injection interfaces.
type Deps struct {
	Input        types.InputCapturer
	UI           types.UIManager
	ScreenWidth  float64
	ScreenHeight float64
	RequestState func(types.GameState) error
	GameState    interface{}
	Debug        DebugConfig

	// DefaultFrameTime is the engine's fallback seconds-per-frame for animated
	// sprites (config.Settings.Animation.DefaultFrameTime), for states/games
	// that want the engine default without importing pkg/config themselves.
	DefaultFrameTime float64
}

// State is a top-level game state. Each State owns its own ecs.World; the engine
// renders the World and drives Update while the State is active.
type State interface {
	// Name returns the state identifier (for logging/debugging).
	Name() string

	// World returns the ECS world backing this state. The engine iterates it to
	// render, and systems operate on it during Update.
	World() *ecs.World

	// Enter activates the state: store deps, build entities, register systems.
	Enter(deps Deps) error

	// Update advances the state's systems by dt seconds.
	Update(dt float64)

	// Exit tears the state down (drop entities/resources).
	Exit()
}

// OverlayRenderer is an optional interface a State may implement to draw
// screen-space overlays (menus, HUD, debug console) after world rendering.
// Replaces SceneOverlayRenderer.
type OverlayRenderer interface {
	DrawOverlays() error
}

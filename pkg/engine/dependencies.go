//go:build js

package engine

import (
	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// EngineDependencies holds all injectable dependencies from the engine.
// This struct is passed to scenes that implement the SceneInjectable interface,
// allowing them to receive all engine services in a single call rather than
// through multiple setter methods.
//
// This pattern:
//   - Reduces boilerplate (42 lines of manual injection → 5 lines)
//   - Makes dependencies explicit and discoverable
//   - Simplifies testing (can create mock dependencies easily)
//   - Provides a single injection point for future extensions
type EngineDependencies struct {
	// InputCapturer provides access to keyboard and gamepad input state
	InputCapturer types.InputCapturer

	// CanvasManager provides access to the WebGPU rendering system
	CanvasManager canvas.CanvasManager

	// StateChangeCallback allows scenes to request pipeline state changes
	// (e.g., transitioning from gameplay to menu)
	StateChangeCallback func(state types.GameState) error

	// GameStateProvider is the game's global state manager
	// The type is defined by the game, not the engine (interface{})
	GameStateProvider interface{}

	// ScreenWidth is the virtual game resolution width
	ScreenWidth float64

	// ScreenHeight is the virtual game resolution height
	ScreenHeight float64
}

// GetDependencies creates and returns a dependency container with all available engine services.
// This is called by the engine when initializing scenes that implement SceneInjectable.
func (e *Engine) GetDependencies() *EngineDependencies {
	return &EngineDependencies{
		InputCapturer:       e.inputCapturer,
		CanvasManager:       e.canvasManager,
		StateChangeCallback: e.SetGameState,
		GameStateProvider:   e.gameStateProvider,
		ScreenWidth:         e.screenWidth,
		ScreenHeight:        e.screenHeight,
	}
}

// Implement types.DependencyProvider interface

// GetInputCapturer returns the input capturer
func (d *EngineDependencies) GetInputCapturer() types.InputCapturer {
	return d.InputCapturer
}

// GetCanvasManager returns the canvas manager as interface{}
func (d *EngineDependencies) GetCanvasManager() interface{} {
	return d.CanvasManager
}

// GetStateChangeCallback returns the state change callback
func (d *EngineDependencies) GetStateChangeCallback() func(types.GameState) error {
	return d.StateChangeCallback
}

// GetGameStateProvider returns the game state provider
func (d *EngineDependencies) GetGameStateProvider() interface{} {
	return d.GameStateProvider
}

// GetScreenWidth returns the screen width
func (d *EngineDependencies) GetScreenWidth() float64 {
	return d.ScreenWidth
}

// GetScreenHeight returns the screen height
func (d *EngineDependencies) GetScreenHeight() float64 {
	return d.ScreenHeight
}


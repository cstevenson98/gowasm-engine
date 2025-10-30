//go:build js

package input

import (
	"sync"
	"syscall/js"

	"github.com/conor/webgpu-triangle/pkg/logger"
	"github.com/conor/webgpu-triangle/pkg/types"
)

// GamepadInput captures gamepad/controller input from the browser
type GamepadInput struct {
	inputState       types.InputState
	mu               sync.RWMutex
	gamepadIndex     int
	hasGamepad       bool
	connectedFunc    js.Func
	disconnectedFunc js.Func
	initialized      bool
	deadzone         float64 // Analog stick deadzone
}

// NewGamepadInput creates a new gamepad input capturer
func NewGamepadInput() *GamepadInput {
	return &GamepadInput{
		inputState:   types.InputState{},
		gamepadIndex: -1,
		hasGamepad:   false,
		initialized:  false,
		deadzone:     0.2, // 20% deadzone for analog sticks
	}
}

// GetInputState returns the current input state
func (g *GamepadInput) GetInputState() types.InputState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.inputState
}

// Initialize sets up gamepad event listeners and checks for existing gamepads
func (g *GamepadInput) Initialize() error {
	if g.initialized {
		return nil
	}

	// Create gamepad connected handler
	g.connectedFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		event := args[0]
		gamepad := event.Get("gamepad")
		index := gamepad.Get("index").Int()
		id := gamepad.Get("id").String()

		g.mu.Lock()
		g.gamepadIndex = index
		g.hasGamepad = true
		g.mu.Unlock()

		logger.Logger.Debugf("Gamepad connected - Index: %d ID: %s", index, id)
		return nil
	})

	// Create gamepad disconnected handler
	g.disconnectedFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		event := args[0]
		gamepad := event.Get("gamepad")
		index := gamepad.Get("index").Int()

		g.mu.Lock()
		if g.gamepadIndex == index {
			g.hasGamepad = false
			g.gamepadIndex = -1
			// Reset input state
			g.inputState = types.InputState{}
		}
		g.mu.Unlock()

		logger.Logger.Debugf("Gamepad disconnected - Index: %d", index)
		return nil
	})

	// Add event listeners
	window := js.Global().Get("window")
	window.Call("addEventListener", "gamepadconnected", g.connectedFunc)
	window.Call("addEventListener", "gamepaddisconnected", g.disconnectedFunc)

	// Check for already connected gamepads
	g.checkExistingGamepads()

	g.initialized = true
	logger.Logger.Debugf("Gamepad input initialized")

	return nil
}

// checkExistingGamepads checks if any gamepads are already connected
func (g *GamepadInput) checkExistingGamepads() {
	navigator := js.Global().Get("navigator")
	gamepads := navigator.Call("getGamepads")

	if gamepads.IsNull() || gamepads.IsUndefined() {
		return
	}

	length := gamepads.Length()
	for i := 0; i < length; i++ {
		gamepad := gamepads.Index(i)
		if !gamepad.IsNull() && !gamepad.IsUndefined() {
			g.mu.Lock()
			g.gamepadIndex = i
			g.hasGamepad = true
			g.mu.Unlock()

			id := gamepad.Get("id").String()
			logger.Logger.Debugf("Found existing gamepad - Index: %d ID: %s", i, id)
			return
		}
	}
}

// Update polls the gamepad state and updates input state
// This should be called every frame
func (g *GamepadInput) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.hasGamepad || g.gamepadIndex < 0 {
		return
	}

	// Get the current gamepad state
	navigator := js.Global().Get("navigator")
	gamepads := navigator.Call("getGamepads")

	if gamepads.IsNull() || gamepads.IsUndefined() {
		return
	}

	gamepad := gamepads.Index(g.gamepadIndex)
	if gamepad.IsNull() || gamepad.IsUndefined() {
		return
	}

	// Read axes (typically: 0=left stick X, 1=left stick Y, 2=right stick X, 3=right stick Y)
	axes := gamepad.Get("axes")
	if !axes.IsNull() && !axes.IsUndefined() && axes.Length() >= 2 {
		leftStickX := axes.Index(0).Float()
		leftStickY := axes.Index(1).Float()

		// Apply deadzone
		if leftStickX < -g.deadzone {
			g.inputState.MoveLeft = true
		} else if leftStickX > g.deadzone {
			g.inputState.MoveRight = true
		} else {
			g.inputState.MoveLeft = false
			g.inputState.MoveRight = false
		}

		if leftStickY < -g.deadzone {
			g.inputState.MoveUp = true
		} else if leftStickY > g.deadzone {
			g.inputState.MoveDown = true
		} else {
			g.inputState.MoveUp = false
			g.inputState.MoveDown = false
		}
	}

	// Read buttons (D-pad: typically buttons 12-15)
	buttons := gamepad.Get("buttons")
	if !buttons.IsNull() && !buttons.IsUndefined() {
		length := buttons.Length()

		// D-pad Up (button 12)
		if length > 12 {
			pressed := buttons.Index(12).Get("pressed").Bool()
			if pressed {
				g.inputState.MoveUp = true
			}
		}

		// D-pad Down (button 13)
		if length > 13 {
			pressed := buttons.Index(13).Get("pressed").Bool()
			if pressed {
				g.inputState.MoveDown = true
			}
		}

		// D-pad Left (button 14)
		if length > 14 {
			pressed := buttons.Index(14).Get("pressed").Bool()
			if pressed {
				g.inputState.MoveLeft = true
			}
		}

		// D-pad Right (button 15)
		if length > 15 {
			pressed := buttons.Index(15).Get("pressed").Bool()
			if pressed {
				g.inputState.MoveRight = true
			}
		}
	}
}

// HasGamepad returns true if a gamepad is currently connected
func (g *GamepadInput) HasGamepad() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hasGamepad
}

// Cleanup releases gamepad event listeners
func (g *GamepadInput) Cleanup() {
	if !g.initialized {
		return
	}

	window := js.Global().Get("window")
	window.Call("removeEventListener", "gamepadconnected", g.connectedFunc)
	window.Call("removeEventListener", "gamepaddisconnected", g.disconnectedFunc)

	g.connectedFunc.Release()
	g.disconnectedFunc.Release()

	g.initialized = false
	logger.Logger.Debugf("Gamepad input cleaned up")
}

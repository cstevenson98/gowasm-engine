//go:build js

package input

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// UnifiedInput combines keyboard and gamepad input into a single input capturer
// It automatically detects and uses gamepad input when available, while also
// supporting keyboard input at all times
type UnifiedInput struct {
	keyboard *KeyboardInput
	gamepad  *GamepadInput
	mu       sync.RWMutex
}

// NewUnifiedInput creates a new unified input capturer
func NewUnifiedInput() *UnifiedInput {
	return &UnifiedInput{
		keyboard: NewKeyboardInput(),
		gamepad:  NewGamepadInput(),
	}
}

// GetInputState returns the combined input state from keyboard and gamepad
// If either input method indicates movement, it will be reflected in the state
func (u *UnifiedInput) GetInputState() types.InputState {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Update gamepad state (must be polled each frame)
	u.gamepad.Update()

	// Get states from both input sources
	keyboardState := u.keyboard.GetInputState()
	gamepadState := u.gamepad.GetInputState()

	// Combine them (OR operation - if either is pressed, it's pressed)
	return types.InputState{
		MoveUp:    keyboardState.MoveUp || gamepadState.MoveUp,
		MoveDown:  keyboardState.MoveDown || gamepadState.MoveDown,
		MoveLeft:  keyboardState.MoveLeft || gamepadState.MoveLeft,
		MoveRight: keyboardState.MoveRight || gamepadState.MoveRight,

		// Arrow keys (keyboard only for menu navigation)
		UpPressed:    keyboardState.UpPressed,
		DownPressed:  keyboardState.DownPressed,
		LeftPressed:  keyboardState.LeftPressed,
		RightPressed: keyboardState.RightPressed,

		// Action keys (keyboard only)
		EnterPressed: keyboardState.EnterPressed,
		SpacePressed: keyboardState.SpacePressed,
		F2Pressed:    keyboardState.F2Pressed,
		Key1Pressed:  keyboardState.Key1Pressed,
		Key2Pressed:  keyboardState.Key2Pressed,

		// Modifier keys (keyboard only)
		ShiftPressed: keyboardState.ShiftPressed,

		// Previous frame states
		UpPressedLastFrame:    keyboardState.UpPressedLastFrame,
		DownPressedLastFrame:  keyboardState.DownPressedLastFrame,
		LeftPressedLastFrame:  keyboardState.LeftPressedLastFrame,
		RightPressedLastFrame: keyboardState.RightPressedLastFrame,
		EnterPressedLastFrame: keyboardState.EnterPressedLastFrame,
		SpacePressedLastFrame: keyboardState.SpacePressedLastFrame,
		F2PressedLastFrame:    keyboardState.F2PressedLastFrame,
		Key1PressedLastFrame:  keyboardState.Key1PressedLastFrame,
		Key2PressedLastFrame:  keyboardState.Key2PressedLastFrame,
		ShiftPressedLastFrame: keyboardState.ShiftPressedLastFrame,
	}
}

// Initialize sets up both keyboard and gamepad input
func (u *UnifiedInput) Initialize() error {
	// Initialize keyboard
	err := u.keyboard.Initialize()
	if err != nil {
		return err
	}

	// Initialize gamepad
	err = u.gamepad.Initialize()
	if err != nil {
		// Don't fail if gamepad initialization fails - keyboard still works
		println("WARNING: Gamepad initialization failed:", err.Error())
	}

	println("DEBUG: Unified input initialized (Keyboard + Gamepad)")
	return nil
}

// Cleanup releases both keyboard and gamepad resources
func (u *UnifiedInput) Cleanup() {
	u.keyboard.Cleanup()
	u.gamepad.Cleanup()
	println("DEBUG: Unified input cleaned up")
}

// HasGamepad returns true if a gamepad is currently connected
func (u *UnifiedInput) HasGamepad() bool {
	return u.gamepad.HasGamepad()
}

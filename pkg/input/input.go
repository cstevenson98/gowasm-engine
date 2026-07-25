package input

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Input implements InputCapturer using Ebiten's input APIs.
// It combines keyboard and gamepad input automatically.
type Input struct {
	state         types.InputState
	previousState types.InputState
}

// NewInput creates a new Ebiten input capturer.
func NewInput() *Input {
	return &Input{
		state:         types.InputState{},
		previousState: types.InputState{},
	}
}

// Initialize sets up input handling (no-op for Ebiten)
func (e *Input) Initialize() error {
	logger.Logger.Debugf("EbitenInput initialized")
	return nil
}

// GetInputState returns the current input state
func (e *Input) GetInputState() types.InputState {
	return e.state
}

// PollInput updates the input state by polling keyboard and gamepad
func (e *Input) PollInput() types.InputState {
	// Save previous state for edge detection
	e.previousState = e.state

	// Reset current state
	e.state = types.InputState{}

	// Poll keyboard
	e.pollKeyboard()

	// Poll mouse
	e.pollMouse()

	// Poll gamepad (if connected)
	e.pollGamepad()

	// Set "last frame" flags based on previous state
	e.state.UpPressedLastFrame = e.previousState.UpPressed
	e.state.DownPressedLastFrame = e.previousState.DownPressed
	e.state.LeftPressedLastFrame = e.previousState.LeftPressed
	e.state.RightPressedLastFrame = e.previousState.RightPressed
	e.state.EnterPressedLastFrame = e.previousState.EnterPressed
	e.state.SpacePressedLastFrame = e.previousState.SpacePressed
	e.state.Key1PressedLastFrame = e.previousState.Key1Pressed
	e.state.Key2PressedLastFrame = e.previousState.Key2Pressed
	e.state.Key3PressedLastFrame = e.previousState.Key3Pressed
	e.state.MPressedLastFrame = e.previousState.MPressed
	e.state.ShiftPressedLastFrame = e.previousState.ShiftPressed
	e.state.CtrlPressedLastFrame = e.previousState.CtrlPressed
	e.state.Mouse.Left.PressedLastFrame = e.previousState.Mouse.Left.Pressed

	return e.state
}

// pollKeyboard polls keyboard input
func (e *Input) pollKeyboard() {
	// Movement keys (continuous)
	// WASD + Arrow keys
	moveUp := ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	moveDown := ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	moveLeft := ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	moveRight := ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)

	e.state.MoveUp = moveUp
	e.state.MoveDown = moveDown
	e.state.MoveLeft = moveLeft
	e.state.MoveRight = moveRight

	// Arrow keys (for menu navigation) - current state
	e.state.UpPressed = ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	e.state.DownPressed = ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS)
	e.state.LeftPressed = ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	e.state.RightPressed = ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)

	// Action keys
	e.state.EnterPressed = ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyNumpadEnter)
	e.state.SpacePressed = ebiten.IsKeyPressed(ebiten.KeySpace)
	e.state.Key1Pressed = ebiten.IsKeyPressed(ebiten.Key1)
	e.state.Key2Pressed = ebiten.IsKeyPressed(ebiten.Key2)
	e.state.Key3Pressed = ebiten.IsKeyPressed(ebiten.Key3)
	e.state.MPressed = ebiten.IsKeyPressed(ebiten.KeyM)

	// Modifier keys
	e.state.ShiftPressed = ebiten.IsKeyPressed(ebiten.KeyShift)
	e.state.CtrlPressed = ebiten.IsKeyPressed(ebiten.KeyControl)
}

// pollMouse polls the cursor position and button state. CursorPosition is
// reported by Ebiten in the same virtual coordinate space used by Layout (and
// thus by draw calls), so no manual scaling is needed here.
func (e *Input) pollMouse() {
	x, y := ebiten.CursorPosition()
	e.state.Mouse.X = float64(x)
	e.state.Mouse.Y = float64(y)
	e.state.Mouse.Left.Pressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// pollGamepad polls gamepad input and merges with keyboard state
func (e *Input) pollGamepad() {
	// Get connected gamepads
	gamepadIDs := ebiten.AppendGamepadIDs(nil)
	if len(gamepadIDs) == 0 {
		return // No gamepad connected
	}

	// Use first connected gamepad
	gid := gamepadIDs[0]

	// Check if this is a standard gamepad layout
	if !ebiten.IsStandardGamepadLayoutAvailable(gid) {
		logger.Logger.Tracef("Non-standard gamepad layout detected")
		return
	}

	// D-Pad / Left stick for movement (continuous)
	up := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftTop)
	down := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftBottom)
	left := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftLeft)
	right := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonLeftRight)

	// Also check left analog stick
	axisX := ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickHorizontal)
	axisY := ebiten.StandardGamepadAxisValue(gid, ebiten.StandardGamepadAxisLeftStickVertical)

	deadzone := 0.3
	if axisX < -deadzone {
		left = true
	}
	if axisX > deadzone {
		right = true
	}
	if axisY < -deadzone {
		up = true
	}
	if axisY > deadzone {
		down = true
	}

	// Merge with keyboard state (OR operation)
	e.state.MoveUp = e.state.MoveUp || up
	e.state.MoveDown = e.state.MoveDown || down
	e.state.MoveLeft = e.state.MoveLeft || left
	e.state.MoveRight = e.state.MoveRight || right

	// Menu navigation (discrete presses)
	e.state.UpPressed = e.state.UpPressed || up
	e.state.DownPressed = e.state.DownPressed || down
	e.state.LeftPressed = e.state.LeftPressed || left
	e.state.RightPressed = e.state.RightPressed || right

	// Face buttons
	// A button (bottom) = Enter/Confirm
	aButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightBottom)
	e.state.EnterPressed = e.state.EnterPressed || aButton

	// B button (right) = Space/Cancel
	bButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightRight)
	e.state.SpacePressed = e.state.SpacePressed || bButton

	// X button (left) = Key1
	xButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightLeft)
	e.state.Key1Pressed = e.state.Key1Pressed || xButton

	// Y button (top) = Key2
	yButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonRightTop)
	e.state.Key2Pressed = e.state.Key2Pressed || yButton

	// Start button = debug console toggle (key "3")
	startButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonCenterRight)
	e.state.Key3Pressed = e.state.Key3Pressed || startButton

	// Select/Back button = M (menu)
	selectButton := ebiten.IsStandardGamepadButtonPressed(gid, ebiten.StandardGamepadButtonCenterLeft)
	e.state.MPressed = e.state.MPressed || selectButton
}

// Update is called each frame to poll input (alias for PollInput)
func (e *Input) Update() {
	e.PollInput()
}

// Cleanup releases input resources (no-op for Ebiten)
func (e *Input) Cleanup() {
	logger.Logger.Debugf("EbitenInput cleanup")
}

package types

// InputState represents the current state of user input
type InputState struct {
	// Movement directions (WASD)
	MoveUp    bool
	MoveDown  bool
	MoveLeft  bool
	MoveRight bool

	// Arrow keys for menu navigation
	UpPressed    bool
	DownPressed  bool
	LeftPressed  bool
	RightPressed bool

	// Action keys
	EnterPressed bool
	SpacePressed bool
	Key1Pressed  bool // Key 1 for scene switching
	Key2Pressed  bool // Key 2 for scene switching
	Key3Pressed  bool // Key 3 for debug console toggle
	MPressed     bool // M key for player menu

	// Modifier keys
	ShiftPressed bool
	CtrlPressed  bool // Ctrl key for save shortcuts (Ctrl+S)

	// Mouse state (position + buttons). See MouseState.
	Mouse MouseState

	// Previous frame state for detecting key presses
	UpPressedLastFrame    bool
	DownPressedLastFrame  bool
	LeftPressedLastFrame  bool
	RightPressedLastFrame bool
	EnterPressedLastFrame bool
	SpacePressedLastFrame bool
	Key1PressedLastFrame  bool
	Key2PressedLastFrame  bool
	Key3PressedLastFrame  bool
	MPressedLastFrame     bool
	ShiftPressedLastFrame bool
	CtrlPressedLastFrame  bool
}

// MouseButtonState tracks a single mouse button's current and
// previous-frame pressed state, using the same edge-detection shape as the
// keyboard's Pressed/PressedLastFrame pairs (e.g. Key2Pressed /
// Key2PressedLastFrame), but as a reusable type instead of a duplicated pair
// of flat fields per button.
type MouseButtonState struct {
	Pressed          bool
	PressedLastFrame bool
}

// MouseState is the mouse's per-frame snapshot: cursor position in virtual
// screen-space pixels (the same coordinate space draw calls use), plus
// button states. Only Left is populated today; Right/Middle or a scroll
// delta are natural additions later without changing InputState's shape.
type MouseState struct {
	X, Y float64
	Left MouseButtonState
}

// InputCapturer is the interface for capturing user input
type InputCapturer interface {
	// GetInputState returns the current input state
	GetInputState() InputState

	// Initialize sets up input listeners
	Initialize() error

	// Cleanup releases input resources
	Cleanup()
}

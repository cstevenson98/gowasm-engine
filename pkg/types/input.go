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
	F2Pressed    bool // F2 key for debug console toggle
	Key1Pressed  bool // Key 1 for scene switching
	Key2Pressed  bool // Key 2 for scene switching

	// Modifier keys
	ShiftPressed bool

	// Previous frame state for detecting key presses
	UpPressedLastFrame    bool
	DownPressedLastFrame  bool
	LeftPressedLastFrame  bool
	RightPressedLastFrame bool
	EnterPressedLastFrame bool
	SpacePressedLastFrame bool
	F2PressedLastFrame    bool
	Key1PressedLastFrame  bool
	Key2PressedLastFrame  bool
	ShiftPressedLastFrame bool
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

package types

// InputState represents the current state of user input
type InputState struct {
	// Movement directions (WASD)
	MoveUp    bool
	MoveDown  bool
	MoveLeft  bool
	MoveRight bool

	// Additional keys can be added here
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

package input

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/types"
)

// MockInput is a mock implementation of InputCapturer for testing
type MockInput struct {
	state       types.InputState
	mu          sync.RWMutex
	initialized bool
}

// NewMockInput creates a new mock input capturer
func NewMockInput() *MockInput {
	return &MockInput{
		state:       types.InputState{},
		initialized: false,
	}
}

// GetInputState returns the current input state
func (m *MockInput) GetInputState() types.InputState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// SetInputState sets the input state (for testing)
func (m *MockInput) SetInputState(state types.InputState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
}

// Initialize sets up the mock input
func (m *MockInput) Initialize() error {
	m.initialized = true
	return nil
}

// Cleanup cleans up the mock input
func (m *MockInput) Cleanup() {
	m.initialized = false
	m.state = types.InputState{}
}

// IsInitialized returns whether the mock is initialized (for testing)
func (m *MockInput) IsInitialized() bool {
	return m.initialized
}

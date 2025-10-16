package mover

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// MockMover is a mock implementation of Mover for testing
type MockMover struct {
	position     types.Vector2
	velocity     types.Vector2
	updateCalled bool
	updateCount  int
}

// NewMockMover creates a new mock mover
func NewMockMover(position, velocity types.Vector2) *MockMover {
	return &MockMover{
		position:     position,
		velocity:     velocity,
		updateCalled: false,
		updateCount:  0,
	}
}

// GetPosition returns the current position
func (m *MockMover) GetPosition() types.Vector2 {
	return m.position
}

// SetPosition sets the position
func (m *MockMover) SetPosition(pos types.Vector2) {
	m.position = pos
}

// GetVelocity returns the current velocity
func (m *MockMover) GetVelocity() types.Vector2 {
	return m.velocity
}

// SetVelocity sets the velocity
func (m *MockMover) SetVelocity(vel types.Vector2) {
	m.velocity = vel
}

// Update updates the position (simple version for testing)
func (m *MockMover) Update(deltaTime float64) {
	m.updateCalled = true
	m.updateCount++
	m.position.X += m.velocity.X * deltaTime
	m.position.Y += m.velocity.Y * deltaTime
}

// SetScreenBounds is a no-op for the mock
func (m *MockMover) SetScreenBounds(width, height float64) {
	// No-op for mock
}

// WasUpdateCalled returns whether Update was called (for testing)
func (m *MockMover) WasUpdateCalled() bool {
	return m.updateCalled
}

// GetUpdateCount returns how many times Update was called (for testing)
func (m *MockMover) GetUpdateCount() int {
	return m.updateCount
}

// ResetUpdateTracking resets the update tracking (for testing)
func (m *MockMover) ResetUpdateTracking() {
	m.updateCalled = false
	m.updateCount = 0
}

package sprite

import (
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MockSprite is a mock implementation of Sprite for testing
type MockSprite struct {
	texturePath  string
	size         types.Vector2
	visible      bool
	updateCalled bool
	updateCount  int
}

// NewMockSprite creates a new mock sprite
func NewMockSprite(texturePath string, size types.Vector2) *MockSprite {
	return &MockSprite{
		texturePath:  texturePath,
		size:         size,
		visible:      true,
		updateCalled: false,
		updateCount:  0,
	}
}

// GetTexturePath returns the mock sprite's texture path.
func (m *MockSprite) GetTexturePath() string {
	return m.texturePath
}

// GetUV returns a full-texture UV rectangle.
func (m *MockSprite) GetUV() types.UVRect {
	return types.UVRect{U: 0.0, V: 0.0, W: 1.0, H: 1.0}
}

// GetSize returns the sprite's size
func (m *MockSprite) GetSize() types.Vector2 {
	return m.size
}

// Update updates the sprite (tracks that it was called)
func (m *MockSprite) Update(deltaTime float64) {
	m.updateCalled = true
	m.updateCount++
}

// SetVisible sets visibility
func (m *MockSprite) SetVisible(visible bool) {
	m.visible = visible
}

// IsVisible returns visibility
func (m *MockSprite) IsVisible() bool {
	return m.visible
}

// WasUpdateCalled returns whether Update was called (for testing)
func (m *MockSprite) WasUpdateCalled() bool {
	return m.updateCalled
}

// GetUpdateCount returns how many times Update was called (for testing)
func (m *MockSprite) GetUpdateCount() int {
	return m.updateCount
}

// ResetUpdateTracking resets the update tracking (for testing)
func (m *MockSprite) ResetUpdateTracking() {
	m.updateCalled = false
	m.updateCount = 0
}

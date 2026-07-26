package canvas

import (
	"fmt"

	"github.com/cstevenson98/milo/pkg/types"
)

// MockCanvasManager implements CanvasManager for testing.
type MockCanvasManager struct {
	initialized    bool
	cleanupCalled  bool
	loadedTextures map[string]bool
	drawnRectCount int
}

// NewMockCanvasManager creates a new mock canvas manager.
func NewMockCanvasManager() *MockCanvasManager {
	return &MockCanvasManager{
		loadedTextures: make(map[string]bool),
	}
}

// Initialize simulates canvas initialization.
func (m *MockCanvasManager) Initialize(canvasID string) error {
	m.initialized = true
	return nil
}

// Cleanup simulates resource cleanup.
func (m *MockCanvasManager) Cleanup() error {
	m.cleanupCalled = true
	m.initialized = false
	return nil
}

// LoadTexture records a loaded texture path.
func (m *MockCanvasManager) LoadTexture(path string) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}
	m.loadedTextures[path] = true
	return nil
}

// DrawTexturedRect records a draw call.
func (m *MockCanvasManager) DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}
	if !m.loadedTextures[texturePath] {
		return &CanvasError{Message: fmt.Sprintf("Texture not loaded: %s", texturePath)}
	}
	m.drawnRectCount++
	return nil
}

// DrawColoredRect records a colored-rect draw call.
func (m *MockCanvasManager) DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error {
	if !m.initialized {
		return &CanvasError{Message: "Canvas not initialized"}
	}
	m.drawnRectCount++
	return nil
}

// IsInitialized reports whether the mock is initialized.
func (m *MockCanvasManager) IsInitialized() bool {
	return m.initialized
}

// WasCleanupCalled reports whether Cleanup was called.
func (m *MockCanvasManager) WasCleanupCalled() bool {
	return m.cleanupCalled
}

// DrawnRectCount returns the number of DrawTexturedRect calls.
func (m *MockCanvasManager) DrawnRectCount() int {
	return m.drawnRectCount
}

package canvas

import (
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// CanvasManager defines the interface for managing canvas operations.
// The game-facing API is intentionally small: load a texture by path and
// draw a textured rectangle. Backend-specific concerns (render targets,
// batching, pipelines) are handled internally by the implementation.
type CanvasManager interface {
	// Initialize sets up the canvas and returns success status
	Initialize(canvasID string) error

	// Cleanup releases resources
	Cleanup() error

	// LoadTexture loads a texture from the given path
	LoadTexture(path string) error

	// DrawTexturedRect draws a region (uv) of the texture at position/size
	DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error

	// DrawColoredRect draws a solid colored rectangle (used for UI/debug overlays)
	DrawColoredRect(position types.Vector2, size types.Vector2, color [4]float32) error
}

// CanvasError represents a canvas-related error
type CanvasError struct {
	Message string
}

func (e *CanvasError) Error() string {
	return e.Message
}

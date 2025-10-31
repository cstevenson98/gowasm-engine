//go:build js

package scene

import (
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// SceneLayer represents the fixed rendering layers for 2D sprites
type SceneLayer int

const (
	// BACKGROUND layer - rendered first (back)
	BACKGROUND SceneLayer = iota
	// ENTITIES layer - rendered second (middle)
	ENTITIES
	// UI layer - rendered last (front)
	UI
)

// String returns the string representation of the scene layer
func (s SceneLayer) String() string {
	switch s {
	case BACKGROUND:
		return "BACKGROUND"
	case ENTITIES:
		return "ENTITIES"
	case UI:
		return "UI"
	default:
		return "UNKNOWN"
	}
}

// Scene represents a game scene that manages game objects organized by layers
type Scene interface {
	// Initialize sets up the scene and its resources
	Initialize() error

	// Update updates all game objects in the scene
	Update(deltaTime float64)

	// GetRenderables returns all game objects in the correct render order (layer order)
	GetRenderables() []types.GameObject

	// Cleanup releases scene resources
	Cleanup()

	// GetName returns the scene identifier
	GetName() string
}

package types

import (
	"fmt"
)

type ObjectState struct {
	ID       string
	Position Vector2
	Visible  bool
	Frame    int
}

// CopyObjectState copies an ObjectState
func CopyObjectState(state ObjectState) ObjectState {
	return ObjectState{
		ID:       state.ID,
		Position: state.Position,
	}
}

// GameObject is the interface that all game objects must implement
type GameObject interface {
	// GetSprite returns the sprite associated with this game object
	GetSprite() Sprite

	// GetMover returns the mover component, or nil if this object doesn't move
	GetMover() Mover

	// Update updates the game object's state and its sprite
	Update(deltaTime float64)

	// GetState returns the game object's current state
	GetState() *ObjectState

	// SetState sets the game object's state
	SetState(state ObjectState)

	// GetID returns the game object's unique identifier
	GetID() string
}

// DebugMessagePoster is an optional interface that callbacks can use to post debug messages
// This is defined here to avoid circular dependencies
type DebugMessagePoster interface {
	PostMessage(source, message string)
}

// globalDebugPoster is a global debug message poster that can be set by the debug package
var globalDebugPoster DebugMessagePoster

// SetGlobalDebugPoster sets the global debug message poster
func SetGlobalDebugPoster(poster DebugMessagePoster) {
	globalDebugPoster = poster
}

// PostDebugMessage posts a debug message from a GameObject
func PostDebugMessage(obj GameObject, format string, args ...interface{}) {
	if globalDebugPoster != nil {
		message := fmt.Sprintf(format, args...)
		globalDebugPoster.PostMessage(obj.GetID(), message)
	}
}

// PostDebugMessageSimple posts a simple debug message with a source string
func PostDebugMessageSimple(source string, format string, args ...interface{}) {
	if globalDebugPoster != nil {
		message := fmt.Sprintf(format, args...)
		globalDebugPoster.PostMessage(source, message)
	}
}

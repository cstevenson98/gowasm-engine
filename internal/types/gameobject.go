package types

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

	// GetSpriteRenderData returns the complete render data for this object
	// This combines data from the sprite and position from the mover (if any)
	GetSpriteRenderData() SpriteRenderData

	// Update updates the game object's state and its sprite
	Update(deltaTime float64)

	// GetState returns the game object's current state
	GetState() *ObjectState

	// SetState sets the game object's state
	SetState(state ObjectState)
}

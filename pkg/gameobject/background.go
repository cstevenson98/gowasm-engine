//go:build js

package gameobject

import (
	"github.com/cstevenson98/gowasm-engine/pkg/sprite"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
	"github.com/google/uuid"
)

// StaticMover is a simple mover that doesn't move (for backgrounds)
type StaticMover struct {
	position types.Vector2
}

func (sm *StaticMover) Update(deltaTime float64) {
	// Static - no movement
}

func (sm *StaticMover) GetPosition() types.Vector2 {
	return sm.position
}

func (sm *StaticMover) SetPosition(pos types.Vector2) {
	sm.position = pos
}

func (sm *StaticMover) GetVelocity() types.Vector2 {
	return types.Vector2{X: 0, Y: 0}
}

func (sm *StaticMover) SetVelocity(vel types.Vector2) {
	// Static - ignore velocity changes
}

func (sm *StaticMover) SetScreenBounds(width, height float64) {
	// Static - no wrapping needed
}

// Background is a GameObject that represents a static background image.
// It embeds BaseGameObject to inherit common GameObject functionality.
type Background struct {
	*BaseGameObject
}

// NewBackground creates a new Background GameObject
// position: top-left corner position
// size: width and height of the background
// texturePath: path to the background texture
func NewBackground(position types.Vector2, size types.Vector2, texturePath string) *Background {
	// Create a single-frame sprite (no animation)
	backgroundSprite := sprite.NewSpriteSheet(
		texturePath,
		sprite.Vector2{X: size.X, Y: size.Y},
		1, // 1 column
		1, // 1 row = 1 frame (static image)
	)

	// Prevent any animation
	backgroundSprite.SetCurrentFrame(0)
	backgroundSprite.SetFrameTime(999999.0) // Extremely long frame time

	// Create a simple mover to provide position (but with zero velocity)
	backgroundMover := &StaticMover{position: position}

	// Create state
	backgroundState := types.ObjectState{
		ID:       uuid.New().String(),
		Position: position,
		Visible:  true,
		Frame:    0,
	}

	// Initialize BaseGameObject
	baseGameObject := NewBaseGameObject(backgroundSprite, backgroundMover, backgroundState)

	return &Background{
		BaseGameObject: baseGameObject,
	}
}

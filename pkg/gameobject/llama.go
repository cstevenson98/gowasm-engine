//go:build js

package gameobject

import (
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/mover"
	"github.com/cstevenson98/gowasm-engine/pkg/sprite"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
	"github.com/google/uuid"
)

// Llama is a GameObject that represents a llama character.
// It embeds BaseGameObject to inherit common GameObject functionality.
type Llama struct {
	*BaseGameObject
}

// NewLlama creates a new Llama GameObject
func NewLlama(position types.Vector2, size types.Vector2, speed float64) *Llama {
	// Create the sprite (just handles texture and animation)
	llamaSprite := sprite.NewSpriteSheet(
		"llama.png",
		sprite.Vector2{X: size.X, Y: size.Y},
		2, // 2 columns (n)
		3, // 3 rows (m) = 6 frames total
	)

	// Set animation speed (slight variation based on speed)
	llamaSprite.SetFrameTime(config.Global.Animation.DefaultFrameTime + (speed/100.0)*0.1)

	// Create the mover (handles position and velocity)
	llamaMover := mover.NewBasicMover(
		position,
		types.Vector2{X: speed, Y: 0}, // Velocity to move right
		size.X,                        // Sprite width for wrapping
		size.Y,                        // Sprite height for wrapping
	)

	// Set screen bounds for wrapping from config
	llamaMover.SetScreenBounds(config.Global.Screen.Width, config.Global.Screen.Height)

	// Create state
	llamaState := types.ObjectState{
		ID:       uuid.New().String(),
		Position: position,
		Visible:  true,
		Frame:    0,
	}

	// Initialize BaseGameObject
	baseGameObject := NewBaseGameObject(llamaSprite, llamaMover, llamaState)

	return &Llama{
		BaseGameObject: baseGameObject,
	}
}

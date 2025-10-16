//go:build js

package gameobject

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
	"github.com/google/uuid"
)

// Llama is a GameObject that represents a llama character
type Llama struct {
	sprite types.Sprite
	state  types.ObjectState

	mu sync.Mutex
}

// NewLlama creates a new Llama GameObject
func NewLlama(position sprite.Vector2, size sprite.Vector2, speed float64) *Llama {
	llamaSprite := sprite.NewSpriteSheet(
		"llama.png",
		position,
		size,
		2, // 2 columns (n)
		3, // 3 rows (m) = 6 frames total
	)

	// Set animation speed
	llamaSprite.SetFrameTime(0.1 + (speed/100.0)*0.1) // Slight variation based on speed

	// Set velocity to move right
	llamaSprite.SetVelocity(sprite.Vector2{X: speed, Y: 0})

	// Set screen bounds for wrapping (these match the defaults but make it explicit)
	llamaSprite.SetScreenBounds(800, 600)

	return &Llama{
		sprite: llamaSprite,
		state: types.ObjectState{
			ID:       uuid.New().String(),
			Position: types.Vector2{X: 0, Y: 0},
			Visible:  true,
			Frame:    0,
		},
	}
}

// GetSprite returns the sprite associated with this Llama
func (l *Llama) GetSprite() types.Sprite {
	return l.sprite
}

// GetState returns the Llama's current state
func (l *Llama) GetState() *types.ObjectState {
	return &l.state
}

// SetState sets the Llama's state
func (l *Llama) SetState(state types.ObjectState) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = types.CopyObjectState(state)
}

// Update updates the Llama's state
func (l *Llama) Update(deltaTime float64) {
	// For now, the sprite handles all updates (animation, movement)
	// In the future, we could add Llama-specific behavior here
	l.sprite.Update(deltaTime)
}

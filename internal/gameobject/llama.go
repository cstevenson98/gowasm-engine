//go:build js

package gameobject

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/mover"
	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
	"github.com/google/uuid"
)

// Llama is a GameObject that represents a llama character
type Llama struct {
	sprite types.Sprite
	mover  types.Mover
	state  types.ObjectState

	mu sync.Mutex
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

	return &Llama{
		sprite: llamaSprite,
		mover:  llamaMover,
		state: types.ObjectState{
			ID:       uuid.New().String(),
			Position: position,
			Visible:  true,
			Frame:    0,
		},
	}
}

// GetSprite returns the sprite associated with this Llama
func (l *Llama) GetSprite() types.Sprite {
	return l.sprite
}

// GetMover returns the mover component for this Llama
func (l *Llama) GetMover() types.Mover {
	return l.mover
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
}

// GetID returns the Llama's unique identifier
func (l *Llama) GetID() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state.ID
}

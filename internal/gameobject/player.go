//go:build js

package gameobject

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/mover"
	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
	"github.com/google/uuid"
)

// Player is a GameObject that represents the player character
type Player struct {
	sprite types.Sprite
	mover  types.Mover
	state  types.ObjectState

	moveSpeed float64 // Base movement speed in pixels per second

	mu sync.Mutex
}

// NewPlayer creates a new Player GameObject
func NewPlayer(position types.Vector2, size types.Vector2, moveSpeed float64) *Player {
	// Create the sprite (just handles texture and animation)
	playerSprite := sprite.NewSpriteSheet(
		"llama.png",
		sprite.Vector2{X: size.X, Y: size.Y},
		2, // 2 columns (n)
		3, // 3 rows (m) = 6 frames total
	)

	// Set animation speed
	playerSprite.SetFrameTime(0.15) // Medium animation speed

	// Create the mover (handles position and velocity)
	playerMover := mover.NewBasicMover(
		position,
		types.Vector2{X: 0, Y: 0}, // Start stationary
		size.X,                    // Sprite width for wrapping
		size.Y,                    // Sprite height for wrapping
	)

	// Set screen bounds for wrapping
	playerMover.SetScreenBounds(800, 600)

	return &Player{
		sprite:    playerSprite,
		mover:     playerMover,
		moveSpeed: moveSpeed,
		state: types.ObjectState{
			ID:       uuid.New().String(),
			Position: position,
			Visible:  true,
			Frame:    0,
		},
	}
}

// GetSprite returns the sprite associated with this Player
func (p *Player) GetSprite() types.Sprite {
	return p.sprite
}

// GetMover returns the mover component for this Player
func (p *Player) GetMover() types.Mover {
	return p.mover
}

// GetState returns the Player's current state
func (p *Player) GetState() *types.ObjectState {
	return &p.state
}

// SetState sets the Player's state
func (p *Player) SetState(state types.ObjectState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = types.CopyObjectState(state)
}

// Update updates the Player's state
func (p *Player) Update(deltaTime float64) {
	// Player-specific behavior can go here
	// (Movement is handled externally by input system updating the mover's velocity)
}

// HandleInput updates the player's velocity based on input state
func (p *Player) HandleInput(inputState types.InputState) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate velocity based on input
	var velocityX, velocityY float64

	if inputState.MoveLeft {
		velocityX -= p.moveSpeed
	}
	if inputState.MoveRight {
		velocityX += p.moveSpeed
	}
	if inputState.MoveUp {
		velocityY -= p.moveSpeed
	}
	if inputState.MoveDown {
		velocityY += p.moveSpeed
	}

	// Normalize diagonal movement to maintain consistent speed
	if velocityX != 0 && velocityY != 0 {
		// Diagonal movement - reduce by sqrt(2) to maintain speed
		velocityX *= 0.7071 // 1/sqrt(2)
		velocityY *= 0.7071
	}

	// Update the mover's velocity
	p.mover.SetVelocity(types.Vector2{X: velocityX, Y: velocityY})
}

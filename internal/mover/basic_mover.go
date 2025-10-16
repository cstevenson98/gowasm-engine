package mover

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// BasicMover implements the Mover interface with screen wrapping
type BasicMover struct {
	position     types.Vector2
	velocity     types.Vector2
	screenWidth  float64
	screenHeight float64
	spriteWidth  float64 // Width of the sprite for wrapping calculations
	spriteHeight float64 // Height of the sprite for wrapping calculations
}

// NewBasicMover creates a new BasicMover
func NewBasicMover(position types.Vector2, velocity types.Vector2, spriteWidth, spriteHeight float64) *BasicMover {
	return &BasicMover{
		position:     position,
		velocity:     velocity,
		screenWidth:  800, // Default screen width
		screenHeight: 600, // Default screen height
		spriteWidth:  spriteWidth,
		spriteHeight: spriteHeight,
	}
}

// GetPosition returns the current position
func (m *BasicMover) GetPosition() types.Vector2 {
	return m.position
}

// SetPosition sets the position
func (m *BasicMover) SetPosition(pos types.Vector2) {
	m.position = pos
}

// GetVelocity returns the current velocity
func (m *BasicMover) GetVelocity() types.Vector2 {
	return m.velocity
}

// SetVelocity sets the velocity
func (m *BasicMover) SetVelocity(vel types.Vector2) {
	m.velocity = vel
}

// Update updates the position based on velocity and handles screen wrapping
func (m *BasicMover) Update(deltaTime float64) {
	// Update position based on velocity
	m.position.X += m.velocity.X * deltaTime
	m.position.Y += m.velocity.Y * deltaTime

	// Screen wrapping - loop back when going off screen
	if m.velocity.X > 0 { // Moving right
		if m.position.X > m.screenWidth {
			m.position.X = -m.spriteWidth // Wrap to left side (just off screen)
		}
	} else if m.velocity.X < 0 { // Moving left
		if m.position.X+m.spriteWidth < 0 {
			m.position.X = m.screenWidth // Wrap to right side
		}
	}

	if m.velocity.Y > 0 { // Moving down
		if m.position.Y > m.screenHeight {
			m.position.Y = -m.spriteHeight // Wrap to top
		}
	} else if m.velocity.Y < 0 { // Moving up
		if m.position.Y+m.spriteHeight < 0 {
			m.position.Y = m.screenHeight // Wrap to bottom
		}
	}
}

// SetScreenBounds sets the screen boundaries for wrapping
func (m *BasicMover) SetScreenBounds(width, height float64) {
	m.screenWidth = width
	m.screenHeight = height
}

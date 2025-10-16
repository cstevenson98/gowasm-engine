package types

// Mover is an interface for objects that can move
// GameObjects may or may not have a Mover component
type Mover interface {
	// GetPosition returns the current position
	GetPosition() Vector2

	// SetPosition sets the position
	SetPosition(pos Vector2)

	// GetVelocity returns the current velocity (pixels per second)
	GetVelocity() Vector2

	// SetVelocity sets the velocity (pixels per second)
	SetVelocity(vel Vector2)

	// Update updates the position based on velocity and deltaTime
	Update(deltaTime float64)

	// SetScreenBounds sets the screen boundaries for wrapping
	SetScreenBounds(width, height float64)
}

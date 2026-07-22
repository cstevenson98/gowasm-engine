package types

// Sprite describes the visual appearance of a game object.
// Sprites handle texture, animation, and size - NOT position or movement.
// A sprite reports its own appearance; positioning is the caller's concern.
type Sprite interface {
	// GetTexturePath returns the path to the sprite's texture (or sprite sheet).
	GetTexturePath() string

	// GetUV returns the UV rectangle for the current animation frame.
	GetUV() UVRect

	// GetSize returns the sprite's display size.
	GetSize() Vector2

	// Update advances the sprite's animation state.
	Update(deltaTime float64)

	// SetVisible sets whether the sprite should be rendered.
	SetVisible(visible bool)

	// IsVisible returns whether the sprite is visible.
	IsVisible() bool
}

package types

// SpriteRenderData contains all data needed to render a sprite
type SpriteRenderData struct {
	TexturePath string
	Position    Vector2
	Size        Vector2
	UV          UVRect
	Visible     bool
	Frame       int
}

// Sprite is the interface that all sprite types must implement
// Sprites handle texture, animation, and size - NOT position or movement
type Sprite interface {
	// GetSpriteRenderData returns the data needed to render this sprite at a given position
	GetSpriteRenderData(position Vector2) SpriteRenderData

	// GetSize returns the sprite's display size
	GetSize() Vector2

	// Update updates the sprite's state (animation, etc.)
	Update(deltaTime float64)

	// SetVisible sets whether the sprite should be rendered
	SetVisible(visible bool)

	// IsVisible returns whether the sprite is visible
	IsVisible() bool
}

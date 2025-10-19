package sprite

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// SpriteSheet represents an animated sprite sheet with n x m frames
// It only handles texture, animation, and size - not position or movement
type SpriteSheet struct {
	// Texture information
	texturePath string

	// Size
	size Vector2

	// Sprite sheet grid
	columns int // n columns
	rows    int // m rows

	// Animation state
	currentFrame int
	totalFrames  int
	frameTime    float64 // Time per frame in seconds
	elapsed      float64 // Elapsed time for current frame

	// Rendering state
	visible bool
}

// Vector2 is a 2D vector (position, size, etc.)
type Vector2 struct {
	X float64
	Y float64
}

// NewSpriteSheet creates a new sprite sheet
// texturePath: path to the sprite sheet texture
// size: display size of the sprite
// columns: number of columns (n) in the sprite sheet
// rows: number of rows (m) in the sprite sheet
func NewSpriteSheet(texturePath string, size Vector2, columns, rows int) *SpriteSheet {
	return &SpriteSheet{
		texturePath:  texturePath,
		size:         size,
		columns:      columns,
		rows:         rows,
		currentFrame: 0,
		totalFrames:  columns * rows,
		frameTime:    0.1, // Default 10 FPS
		elapsed:      0.0,
		visible:      true,
	}
}

// GetSpriteRenderData returns the data needed to render this sprite at a given position
func (s *SpriteSheet) GetSpriteRenderData(position types.Vector2) types.SpriteRenderData {
	// Calculate UV coordinates for the current frame
	frameWidth := 1.0 / float64(s.columns)
	frameHeight := 1.0 / float64(s.rows)

	frameX := s.currentFrame % s.columns
	frameY := s.currentFrame / s.columns

	uv := types.UVRect{
		U: float64(frameX) * frameWidth,
		V: float64(frameY) * frameHeight,
		W: frameWidth,
		H: frameHeight,
	}

	return types.SpriteRenderData{
		TexturePath: s.texturePath,
		Position:    position,
		Size:        types.Vector2{X: s.size.X, Y: s.size.Y},
		UV:          uv,
		Visible:     s.visible,
	}
}

// GetSize returns the sprite's size
func (s *SpriteSheet) GetSize() types.Vector2 {
	return types.Vector2{X: s.size.X, Y: s.size.Y}
}

// Update updates the sprite's animation state
func (s *SpriteSheet) Update(deltaTime float64) {
	// Update animation
	if s.totalFrames > 1 {
		s.elapsed += deltaTime

		// Advance to next frame if enough time has passed
		if s.elapsed >= s.frameTime {
			s.elapsed -= s.frameTime
			s.currentFrame = (s.currentFrame + 1) % s.totalFrames
		}
	}
	// If totalFrames == 1, this is a static sprite - no animation
}

// SetVisible sets whether the sprite should be rendered
func (s *SpriteSheet) SetVisible(visible bool) {
	s.visible = visible
}

// IsVisible returns whether the sprite is visible
func (s *SpriteSheet) IsVisible() bool {
	return s.visible
}

// SetFrameTime sets the time per frame for animation (in seconds)
func (s *SpriteSheet) SetFrameTime(frameTime float64) {
	s.frameTime = frameTime
}

// SetCurrentFrame sets the current frame index (0-based)
func (s *SpriteSheet) SetCurrentFrame(frame int) {
	if frame >= 0 && frame < s.totalFrames {
		s.currentFrame = frame
	}
}

// GetCurrentFrame returns the current frame index
func (s *SpriteSheet) GetCurrentFrame() int {
	return s.currentFrame
}

// GetTotalFrames returns the total number of frames
func (s *SpriteSheet) GetTotalFrames() int {
	return s.totalFrames
}

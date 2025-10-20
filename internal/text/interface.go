package text

import (
	"github.com/conor/webgpu-triangle/internal/types"
)

// Font represents a font sprite sheet with character mapping
type Font interface {
	// GetCharacterUV returns the UV coordinates for a given character
	GetCharacterUV(char rune) (types.UVRect, error)

	// GetTexturePath returns the path to the font sprite sheet texture
	GetTexturePath() string

	// GetCellSize returns the width and height of each character cell
	GetCellSize() (int, int)

	// IsLoaded returns true if the font is loaded and ready to use
	IsLoaded() bool
}

// TextRenderer handles rendering text using font sprite sheets
type TextRenderer interface {
	// RenderText renders a string at the given position with the specified color
	RenderText(text string, position types.Vector2, font Font, color [4]float32) error

	// RenderTextScaled renders a string at the given position with scaling and color
	RenderTextScaled(text string, position types.Vector2, font Font, scale float64, color [4]float32) error
}

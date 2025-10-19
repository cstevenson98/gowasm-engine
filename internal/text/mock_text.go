package text

import (
	"fmt"

	"github.com/conor/webgpu-triangle/internal/types"
)

// MockFont is a mock implementation of the Font interface for testing
type MockFont struct {
	TexturePath    string
	CellW          int
	CellH          int
	LoadedFlag     bool
	CharacterUVMap map[rune]types.UVRect
}

// NewMockFont creates a new mock font
func NewMockFont() *MockFont {
	return &MockFont{
		TexturePath:    "mock_font.png",
		CellW:          16,
		CellH:          16,
		LoadedFlag:     true,
		CharacterUVMap: make(map[rune]types.UVRect),
	}
}

// GetCharacterUV returns the UV coordinates for a given character
func (m *MockFont) GetCharacterUV(char rune) (types.UVRect, error) {
	if !m.LoadedFlag {
		return types.UVRect{}, fmt.Errorf("font not loaded")
	}

	uv, exists := m.CharacterUVMap[char]
	if !exists {
		return types.UVRect{}, fmt.Errorf("character not found: %c", char)
	}

	return uv, nil
}

// GetTexturePath returns the path to the font sprite sheet texture
func (m *MockFont) GetTexturePath() string {
	return m.TexturePath
}

// GetCellSize returns the width and height of each character cell
func (m *MockFont) GetCellSize() (int, int) {
	return m.CellW, m.CellH
}

// IsLoaded returns true if the font is loaded and ready to use
func (m *MockFont) IsLoaded() bool {
	return m.LoadedFlag
}

// MockTextRenderer is a mock implementation of the TextRenderer interface for testing
type MockTextRenderer struct {
	RenderedTexts []MockRenderedText
}

// MockRenderedText represents a text rendering call
type MockRenderedText struct {
	Text     string
	Position types.Vector2
	Font     Font
	Scale    float64
	Color    [4]float32
}

// NewMockTextRenderer creates a new mock text renderer
func NewMockTextRenderer() *MockTextRenderer {
	return &MockTextRenderer{
		RenderedTexts: make([]MockRenderedText, 0),
	}
}

// RenderText renders a string at the given position with the specified color
func (m *MockTextRenderer) RenderText(text string, position types.Vector2, font Font, color [4]float32) error {
	m.RenderedTexts = append(m.RenderedTexts, MockRenderedText{
		Text:     text,
		Position: position,
		Font:     font,
		Scale:    1.0,
		Color:    color,
	})
	return nil
}

// RenderTextScaled renders a string at the given position with scaling and color
func (m *MockTextRenderer) RenderTextScaled(text string, position types.Vector2, font Font, scale float64, color [4]float32) error {
	m.RenderedTexts = append(m.RenderedTexts, MockRenderedText{
		Text:     text,
		Position: position,
		Font:     font,
		Scale:    scale,
		Color:    color,
	})
	return nil
}

// GetRenderedTextCount returns the number of texts rendered
func (m *MockTextRenderer) GetRenderedTextCount() int {
	return len(m.RenderedTexts)
}

// Clear clears the rendered texts history
func (m *MockTextRenderer) Clear() {
	m.RenderedTexts = make([]MockRenderedText, 0)
}


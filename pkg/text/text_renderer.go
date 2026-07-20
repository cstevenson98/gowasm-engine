package text

import (
	"fmt"

	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BasicTextRenderer implements the TextRenderer interface
type BasicTextRenderer struct {
	canvasManager canvas.CanvasManager
}

// NewTextRenderer creates a new text renderer
func NewTextRenderer(canvasManager canvas.CanvasManager) *BasicTextRenderer {
	return &BasicTextRenderer{
		canvasManager: canvasManager,
	}
}

// RenderText renders a string at the given position with the specified color
func (r *BasicTextRenderer) RenderText(text string, position types.Vector2, font Font, color [4]float32) error {
	return r.RenderTextScaled(text, position, font, 1.0, color)
}

// RenderTextScaled renders a string at the given position with scaling and color
func (r *BasicTextRenderer) RenderTextScaled(text string, position types.Vector2, font Font, scale float64, color [4]float32) error {
	if !font.IsLoaded() {
		return fmt.Errorf("font not loaded")
	}

	if len(text) == 0 {
		return nil
	}

	cellWidth, cellHeight := font.GetCellSize()
	if cellWidth == 0 || cellHeight == 0 {
		return fmt.Errorf("invalid font cell size: %dx%d", cellWidth, cellHeight)
	}

	// Glyphs are drawn in virtual (game) pixels. The engine upscales the whole
	// virtual screen to the window, so text layout here is independent of the
	// pixel scale - it only depends on the font scale.
	scaledWidth := float64(cellWidth) * scale
	scaledHeight := float64(cellHeight) * scale

	// Horizontal advance per character: the cell width tightened by the
	// configured spacing reduction (mono-font cells carry built-in padding).
	advance := (float64(cellWidth) - config.Global.Debug.CharacterSpacingReduction) * scale

	// Vertical distance between lines for multi-line strings.
	lineHeight := scaledHeight * config.Global.Rendering.TextLineSpacing

	// Current position for rendering (advances with each character)
	currentX := position.X
	currentY := position.Y

	// Render each character
	for _, char := range text {
		// Handle special characters
		if char == '\n' {
			// Newline: move to next line with proper line spacing
			currentX = position.X
			currentY += lineHeight
			continue
		}

		if char == ' ' {
			currentX += advance
			continue
		}

		// Get UV coordinates for this character
		uv, err := font.GetCharacterUV(char)
		if err != nil {
			logger.Logger.Tracef("Character '%c' not found in font, skipping", char)
			currentX += advance
			continue
		}

		// Draw the character using the canvas
		err = r.canvasManager.DrawTexturedRect(
			font.GetTexturePath(),
			types.Vector2{X: currentX, Y: currentY},
			types.Vector2{X: scaledWidth, Y: scaledHeight},
			uv,
		)

		if err != nil {
			// Texture might not be loaded yet - normal during initial loading.
			currentX += advance
			continue
		}

		currentX += advance
	}

	return nil
}

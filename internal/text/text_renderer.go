//go:build js

package text

import (
	"fmt"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/types"
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

	// Scale the cell size
	// Note: The canvas manager will handle pixel-perfect snapping and upscaling automatically
	// based on the PixelScale setting, so we just apply the font scale here
	scaledWidth := float64(cellWidth) * scale
	scaledHeight := float64(cellHeight) * scale

	// Account for pixel scale when advancing character positions
	// The canvas scales sizes by PixelScale, so we need to advance by the same amount
	pixelScale := 1.0
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		pixelScale = float64(config.Global.Rendering.PixelScale)
	}

	// Actual rendered dimensions (after canvas scaling)
	renderedWidth := scaledWidth * pixelScale
	renderedHeight := scaledHeight * pixelScale

	// Line height includes extra spacing between lines for paragraph text
	lineHeight := renderedHeight * config.Global.Rendering.TextLineSpacing

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
			// Space: advance position with reduced spacing
			spacingReduction := config.Global.Debug.CharacterSpacingReduction * scale * pixelScale
			currentX += renderedWidth - spacingReduction
			continue
		}

		// Get UV coordinates for this character
		uv, err := font.GetCharacterUV(char)
		if err != nil {
			logger.Logger.Tracef("Character '%c' not found in font, skipping", char)
			spacingReduction := config.Global.Debug.CharacterSpacingReduction * scale * pixelScale
			currentX += renderedWidth - spacingReduction
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
			// Texture might not be loaded yet - silently skip and continue
			// This is normal during initial loading
			spacingReduction := config.Global.Debug.CharacterSpacingReduction * scale * pixelScale
			currentX += renderedWidth - spacingReduction
			continue
		}

		// Advance to next character position, reducing spacing by configured amount
		spacingReduction := config.Global.Debug.CharacterSpacingReduction * scale * pixelScale
		currentX += renderedWidth - spacingReduction
	}

	return nil
}

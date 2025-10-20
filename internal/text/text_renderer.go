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
	var scaledWidth, scaledHeight float64

	if config.Global.Rendering.PixelArtMode && config.Global.Rendering.PixelPerfectScaling {
		// Use integer scaling for pixel-perfect rendering
		scaleInt := int(scale + 0.5) // Round to nearest integer
		if scaleInt < 1 {
			scaleInt = 1 // Minimum 1x scaling
		}
		scaledWidth = float64(cellWidth * scaleInt)
		scaledHeight = float64(cellHeight * scaleInt)
	} else {
		// Use fractional scaling for smooth rendering
		scaledWidth = float64(cellWidth) * scale
		scaledHeight = float64(cellHeight) * scale
	}

	// Current position for rendering (advances with each character)
	currentX := position.X
	currentY := position.Y

	// Render each character
	for _, char := range text {
		// Handle special characters
		if char == '\n' {
			// Newline: move to next line
			currentX = position.X
			currentY += scaledHeight
			continue
		}

		if char == ' ' {
			// Space: advance position with reduced spacing
			spacingReduction := config.Global.Debug.CharacterSpacingReduction
			if config.Global.Rendering.PixelArtMode && config.Global.Rendering.PixelPerfectScaling {
				// Use integer scaling for spacing reduction
				scaleInt := int(scale + 0.5)
				if scaleInt < 1 {
					scaleInt = 1
				}
				spacingReduction *= float64(scaleInt)
			} else {
				spacingReduction *= scale
			}
			currentX += scaledWidth - spacingReduction
			continue
		}

		// Get UV coordinates for this character
		uv, err := font.GetCharacterUV(char)
		if err != nil {
			logger.Logger.Tracef("Character '%c' not found in font, skipping", char)
			spacingReduction := config.Global.Debug.CharacterSpacingReduction
			if config.Global.Rendering.PixelArtMode && config.Global.Rendering.PixelPerfectScaling {
				scaleInt := int(scale + 0.5)
				if scaleInt < 1 {
					scaleInt = 1
				}
				spacingReduction *= float64(scaleInt)
			} else {
				spacingReduction *= scale
			}
			currentX += scaledWidth - spacingReduction
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
			spacingReduction := config.Global.Debug.CharacterSpacingReduction
			if config.Global.Rendering.PixelArtMode && config.Global.Rendering.PixelPerfectScaling {
				scaleInt := int(scale + 0.5)
				if scaleInt < 1 {
					scaleInt = 1
				}
				spacingReduction *= float64(scaleInt)
			} else {
				spacingReduction *= scale
			}
			currentX += scaledWidth - spacingReduction
			continue
		}

		// Advance to next character position, reducing spacing by configured amount
		spacingReduction := config.Global.Debug.CharacterSpacingReduction
		if config.Global.Rendering.PixelArtMode && config.Global.Rendering.PixelPerfectScaling {
			scaleInt := int(scale + 0.5)
			if scaleInt < 1 {
				scaleInt = 1
			}
			spacingReduction *= float64(scaleInt)
		} else {
			spacingReduction *= scale
		}
		currentX += scaledWidth - spacingReduction
	}

	return nil
}

package canvas

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Canvas implements CanvasManager using Ebiten.
//
// Ebiten coalesces consecutive DrawImage calls that share the same source
// image internally, so draws are issued immediately to the current render
// target without any manual batching layer.
type Canvas struct {
	// Render target (set by the engine each frame during Draw)
	screen *ebiten.Image

	// Loaded textures (path -> image)
	loadedTextures map[string]*ebiten.Image

	initialized bool
}

// NewCanvas creates a new Ebiten canvas.
func NewCanvas() *Canvas {
	return &Canvas{
		loadedTextures: make(map[string]*ebiten.Image),
	}
}

// SetScreen sets the render target (called by the engine during Draw).
func (c *Canvas) SetScreen(screen *ebiten.Image) {
	c.screen = screen
}

// Initialize sets up the canvas (no-op for Ebiten, screen is passed externally).
func (c *Canvas) Initialize(canvasID string) error {
	logger.Logger.Debugf("Canvas.Initialize called (canvasID: %s)", canvasID)
	c.initialized = true
	return nil
}

// Cleanup releases resources.
func (c *Canvas) Cleanup() error {
	logger.Logger.Debugf("Canvas.Cleanup called")
	c.loadedTextures = make(map[string]*ebiten.Image)
	c.initialized = false
	return nil
}

// LoadTexture loads a texture from the filesystem, caching by path.
func (c *Canvas) LoadTexture(path string) error {
	if _, exists := c.loadedTextures[path]; exists {
		logger.Logger.Debugf("Texture already loaded: %s", path)
		return nil
	}

	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to load texture %s: %v", path, err)
		logger.Logger.Errorf(errMsg)
		return &CanvasError{Message: errMsg}
	}

	c.loadedTextures[path] = img
	logger.Logger.Debugf("Loaded texture: %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

// DrawTexturedRect draws a region (uv) of a loaded texture at the given
// position and size, drawing immediately to the current screen.
func (c *Canvas) DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if c.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	img, exists := c.loadedTextures[texturePath]
	if !exists {
		// Lazy-load on first use so dynamically spawned objects render without
		// needing to be pre-declared. Loaded textures are cached thereafter.
		if err := c.LoadTexture(texturePath); err != nil {
			return err
		}
		img = c.loadedTextures[texturePath]
	}

	texWidth := float64(img.Bounds().Dx())
	texHeight := float64(img.Bounds().Dy())

	x0 := int(uv.U * texWidth)
	y0 := int(uv.V * texHeight)
	x1 := int((uv.U + uv.W) * texWidth)
	y1 := int((uv.V + uv.H) * texHeight)

	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)

	pos := snapToPixel(position)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(size.X/float64(x1-x0), size.Y/float64(y1-y0))
	opts.GeoM.Translate(pos.X, pos.Y)

	if config.Global.Rendering.PixelArtMode {
		opts.Filter = ebiten.FilterNearest
	} else {
		opts.Filter = ebiten.FilterLinear
	}

	c.screen.DrawImage(subImg, opts)
	return nil
}

// DrawColoredRect draws a solid colored rectangle, drawing immediately to the
// current screen. Used for UI/debug overlays.
func (c *Canvas) DrawColoredRect(position types.Vector2, size types.Vector2, colorRGBA [4]float32) error {
	if c.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	rect := ebiten.NewImage(int(size.X), int(size.Y))
	rect.Fill(color.RGBA{
		R: uint8(colorRGBA[0] * 255),
		G: uint8(colorRGBA[1] * 255),
		B: uint8(colorRGBA[2] * 255),
		A: uint8(colorRGBA[3] * 255),
	})

	pos := snapToPixel(position)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(pos.X, pos.Y)
	c.screen.DrawImage(rect, opts)
	return nil
}

// snapToPixel rounds a position to whole virtual pixels when pixel-perfect
// rendering is enabled. Movers keep their smooth sub-pixel position; only the
// rendered position is quantized, which keeps sprites and text crisp and free
// of shimmer as they move (the virtual screen is later upscaled by PixelScale).
func snapToPixel(position types.Vector2) types.Vector2 {
	if !config.Global.Rendering.PixelPerfectScaling {
		return position
	}
	return types.Vector2{
		X: math.Round(position.X),
		Y: math.Round(position.Y),
	}
}

package canvas

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// drawCall represents a deferred draw operation for batch rendering
type drawCall struct {
	texture     *ebiten.Image
	options     *ebiten.DrawImageOptions
	texturePath string
}

// EbitenCanvasManager implements CanvasManager using Ebiten
type EbitenCanvasManager struct {
	// Render target (set by engine during Draw)
	screen *ebiten.Image

	// Loaded textures (path -> image)
	loadedTextures map[string]*ebiten.Image

	// Batch rendering
	batchMode      bool
	batchedCalls   []drawCall
	currentBatch   string // Current texture path for batching

	// Pipelines (dummy for compatibility)
	spritePipeline     *EbitenPipeline
	backgroundPipeline *EbitenPipeline

	// Status
	initialized bool
	statusMsg   string
}

// EbitenPipeline is a dummy pipeline type for interface compatibility
type EbitenPipeline struct {
	valid bool
}

func (p *EbitenPipeline) IsValid() bool {
	return p.valid
}

// NewEbitenCanvasManager creates a new Ebiten canvas manager
func NewEbitenCanvasManager() *EbitenCanvasManager {
	return &EbitenCanvasManager{
		loadedTextures:     make(map[string]*ebiten.Image),
		spritePipeline:     &EbitenPipeline{valid: true},
		backgroundPipeline: &EbitenPipeline{valid: true},
	}
}

// SetScreen sets the render target (called by engine during Draw)
func (e *EbitenCanvasManager) SetScreen(screen *ebiten.Image) {
	e.screen = screen
}

// Initialize sets up the canvas (no-op for Ebiten, screen is passed externally)
func (e *EbitenCanvasManager) Initialize(canvasID string) error {
	logger.Logger.Debugf("EbitenCanvasManager.Initialize called (canvasID: %s)", canvasID)
	e.initialized = true
	e.statusMsg = "Ebiten canvas manager initialized"
	return nil
}

// Render flushes any pending operations (no-op, drawing is immediate)
func (e *EbitenCanvasManager) Render() error {
	// Ebiten handles rendering automatically
	return nil
}

// Cleanup releases resources
func (e *EbitenCanvasManager) Cleanup() error {
	logger.Logger.Debugf("EbitenCanvasManager.Cleanup called")
	// Clear texture cache
	e.loadedTextures = make(map[string]*ebiten.Image)
	e.initialized = false
	return nil
}

// GetStatus returns the current status
func (e *EbitenCanvasManager) GetStatus() (bool, string) {
	return e.initialized, e.statusMsg
}

// SetStatus updates the status
func (e *EbitenCanvasManager) SetStatus(initialized bool, message string) {
	e.initialized = initialized
	e.statusMsg = message
}

// LoadTexture loads a texture from the filesystem
func (e *EbitenCanvasManager) LoadTexture(path string) error {
	// Check if already loaded
	if _, exists := e.loadedTextures[path]; exists {
		logger.Logger.Debugf("Texture already loaded: %s", path)
		return nil
	}

	// Load image from file
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to load texture %s: %v", path, err)
		logger.Logger.Errorf(errMsg)
		return &CanvasError{Message: errMsg}
	}

	e.loadedTextures[path] = img
	logger.Logger.Debugf("Loaded texture: %s (%dx%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

// DrawTexture draws a texture at the specified position with UV coordinates
func (e *EbitenCanvasManager) DrawTexture(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	if e.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	texturePath := texture.GetID()
	img, exists := e.loadedTextures[texturePath]
	if !exists {
		errMsg := fmt.Sprintf("Texture not loaded: %s", texturePath)
		logger.Logger.Errorf(errMsg)
		return &CanvasError{Message: errMsg}
	}

	// Calculate UV coordinates in pixels
	texWidth := float64(img.Bounds().Dx())
	texHeight := float64(img.Bounds().Dy())
	
	x0 := int(uv.U * texWidth)
	y0 := int(uv.V * texHeight)
	x1 := int((uv.U + uv.W) * texWidth)
	y1 := int((uv.V + uv.H) * texHeight)

	// Create subimage for UV rect
	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)

	// Create draw options
	opts := &ebiten.DrawImageOptions{}
	
	// Scale to desired size
	scaleX := size.X / float64(x1-x0)
	scaleY := size.Y / float64(y1-y0)
	opts.GeoM.Scale(scaleX, scaleY)
	
	// Translate to position
	opts.GeoM.Translate(position.X, position.Y)
	
	// Use nearest-neighbor filtering for pixel-perfect rendering
	if config.Global.Rendering.PixelArtMode {
		opts.Filter = ebiten.FilterNearest
	} else {
		opts.Filter = ebiten.FilterLinear
	}

	if e.batchMode {
		// Store for batch rendering
		e.batchedCalls = append(e.batchedCalls, drawCall{
			texture:     subImg,
			options:     opts,
			texturePath: texturePath,
		})
	} else {
		// Draw immediately
		e.screen.DrawImage(subImg, opts)
	}

	return nil
}

// DrawTextureRotated draws a texture with rotation
func (e *EbitenCanvasManager) DrawTextureRotated(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, rotation float64) error {
	if e.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	texturePath := texture.GetID()
	img, exists := e.loadedTextures[texturePath]
	if !exists {
		errMsg := fmt.Sprintf("Texture not loaded: %s", texturePath)
		logger.Logger.Errorf(errMsg)
		return &CanvasError{Message: errMsg}
	}

	// Calculate UV coordinates in pixels
	texWidth := float64(img.Bounds().Dx())
	texHeight := float64(img.Bounds().Dy())
	
	x0 := int(uv.U * texWidth)
	y0 := int(uv.V * texHeight)
	x1 := int((uv.U + uv.W) * texWidth)
	y1 := int((uv.V + uv.H) * texHeight)

	// Create subimage for UV rect
	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)

	// Create draw options
	opts := &ebiten.DrawImageOptions{}
	
	// Scale to desired size
	scaleX := size.X / float64(x1-x0)
	scaleY := size.Y / float64(y1-y0)
	
	// Apply transformations in correct order:
	// 1. Scale
	// 2. Rotate around center
	// 3. Translate to position
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(-size.X/2, -size.Y/2) // Move origin to center
	opts.GeoM.Rotate(rotation)                 // Rotate
	opts.GeoM.Translate(position.X+size.X/2, position.Y+size.Y/2) // Move to final position

	// Filtering
	if config.Global.Rendering.PixelArtMode {
		opts.Filter = ebiten.FilterNearest
	} else {
		opts.Filter = ebiten.FilterLinear
	}

	if e.batchMode {
		e.batchedCalls = append(e.batchedCalls, drawCall{
			texture:     subImg,
			options:     opts,
			texturePath: texturePath,
		})
	} else {
		e.screen.DrawImage(subImg, opts)
	}

	return nil
}

// DrawTextureScaled draws a texture with custom scaling
func (e *EbitenCanvasManager) DrawTextureScaled(texture types.Texture, position types.Vector2, size types.Vector2, uv types.UVRect, scale types.Vector2) error {
	if e.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	texturePath := texture.GetID()
	img, exists := e.loadedTextures[texturePath]
	if !exists {
		errMsg := fmt.Sprintf("Texture not loaded: %s", texturePath)
		logger.Logger.Errorf(errMsg)
		return &CanvasError{Message: errMsg}
	}

	// Calculate UV coordinates in pixels
	texWidth := float64(img.Bounds().Dx())
	texHeight := float64(img.Bounds().Dy())
	
	x0 := int(uv.U * texWidth)
	y0 := int(uv.V * texHeight)
	x1 := int((uv.U + uv.W) * texWidth)
	y1 := int((uv.V + uv.H) * texHeight)

	// Create subimage for UV rect
	subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)

	// Create draw options
	opts := &ebiten.DrawImageOptions{}
	
	// Apply base scale for size
	baseScaleX := size.X / float64(x1-x0)
	baseScaleY := size.Y / float64(y1-y0)
	
	// Apply additional scaling
	opts.GeoM.Scale(baseScaleX*scale.X, baseScaleY*scale.Y)
	
	// Translate to position
	opts.GeoM.Translate(position.X, position.Y)
	
	// Filtering
	if config.Global.Rendering.PixelArtMode {
		opts.Filter = ebiten.FilterNearest
	} else {
		opts.Filter = ebiten.FilterLinear
	}

	if e.batchMode {
		e.batchedCalls = append(e.batchedCalls, drawCall{
			texture:     subImg,
			options:     opts,
			texturePath: texturePath,
		})
	} else {
		e.screen.DrawImage(subImg, opts)
	}

	return nil
}

// DrawColoredRect draws a colored rectangle (for debug console)
func (e *EbitenCanvasManager) DrawColoredRect(position types.Vector2, size types.Vector2, colorRGBA [4]float32) error {
	if e.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}

	// Create a 1x1 white image to use as a base
	rect := ebiten.NewImage(int(size.X), int(size.Y))
	
	// Convert float32 RGBA to color.RGBA
	r := uint8(colorRGBA[0] * 255)
	g := uint8(colorRGBA[1] * 255)
	b := uint8(colorRGBA[2] * 255)
	a := uint8(colorRGBA[3] * 255)
	
	rect.Fill(color.RGBA{R: r, G: g, B: b, A: a})

	// Draw to screen
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(position.X, position.Y)
	e.screen.DrawImage(rect, opts)

	return nil
}

// DrawTexturedRect draws a textured rectangle
func (e *EbitenCanvasManager) DrawTexturedRect(texturePath string, position types.Vector2, size types.Vector2, uv types.UVRect) error {
	// Use the texture interface-based method
	texture := types.NewWebGPUTexture(0, 0, texturePath) // Size doesn't matter, we look it up
	return e.DrawTexture(texture, position, size, uv)
}

// BeginBatch starts batch rendering mode
func (e *EbitenCanvasManager) BeginBatch() error {
	e.batchMode = true
	e.batchedCalls = make([]drawCall, 0, 256) // Pre-allocate for common case
	logger.Logger.Tracef("BeginBatch: Batch mode enabled")
	return nil
}

// EndBatch ends batch rendering and flushes all batched draws
func (e *EbitenCanvasManager) EndBatch() error {
	if !e.batchMode {
		return &CanvasError{Message: "Not in batch mode"}
	}

	// Flush all batched draws
	for _, call := range e.batchedCalls {
		e.screen.DrawImage(call.texture, call.options)
	}

	// Reset batch state
	e.batchMode = false
	e.batchedCalls = nil
	
	logger.Logger.Tracef("EndBatch: Flushed batch")
	return nil
}

// FlushBatch flushes current batch without ending batch mode
func (e *EbitenCanvasManager) FlushBatch() error {
	if !e.batchMode {
		return &CanvasError{Message: "Not in batch mode"}
	}

	// Flush all batched draws
	for _, call := range e.batchedCalls {
		e.screen.DrawImage(call.texture, call.options)
	}

	// Clear batch but stay in batch mode
	e.batchedCalls = e.batchedCalls[:0]
	
	logger.Logger.Tracef("FlushBatch: Partial flush")
	return nil
}

// ClearCanvas clears the screen
func (e *EbitenCanvasManager) ClearCanvas() error {
	if e.screen == nil {
		return &CanvasError{Message: "Screen not set (call SetScreen first)"}
	}
	
	e.screen.Clear()
	return nil
}

// GetSpritePipeline returns the sprite pipeline (dummy for Ebiten)
func (e *EbitenCanvasManager) GetSpritePipeline() types.Pipeline {
	return e.spritePipeline
}

// GetBackgroundPipeline returns the background pipeline (dummy for Ebiten)
func (e *EbitenCanvasManager) GetBackgroundPipeline() types.Pipeline {
	return e.backgroundPipeline
}

// SetPipelines sets active pipelines (no-op for Ebiten)
func (e *EbitenCanvasManager) SetPipelines(pipelines []types.PipelineType) error {
	// Ebiten doesn't have explicit pipelines, this is a no-op
	logger.Logger.Tracef("SetPipelines called (no-op for Ebiten)")
	return nil
}

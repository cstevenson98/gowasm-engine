package types

import "math"

// Vector2 represents a 2D vector
type Vector2 struct {
	X float64
	Y float64
}

// Camera represents a 2D camera with position and zoom.
// Position (X, Y) is an offset in game pixel coordinates from the default view.
// (0, 0) means no offset (default view). Positive X pans right, positive Y pans down.
// Zoom is the zoom level (1.0 = default, 2.0 = 2x zoom in, etc.).
// For pixel-perfect rendering, zoom should be a positive integer.
type Camera struct {
	X    float64 // Camera X offset in game pixels (0 = default)
	Y    float64 // Camera Y offset in game pixels (0 = default)
	Zoom float64 // Zoom level (1.0 = no zoom, 2.0 = 2x, etc.)
}

// DefaultCamera returns a camera with default settings (no pan, no zoom).
func DefaultCamera() Camera {
	return Camera{X: 0, Y: 0, Zoom: 1.0}
}

// GetMatrix computes the camera view-projection matrix for GPU upload.
// The matrix is in column-major order (WGSL mat4x4<f32> layout).
// canvasWidth and canvasHeight are the actual canvas pixel dimensions.
// pixelScale is the game-to-screen pixel ratio.
func (c Camera) GetMatrix(canvasWidth, canvasHeight float64, pixelScale int) [16]float32 {
	z := float32(c.Zoom)
	ps := float64(pixelScale)
	if ps < 1 {
		ps = 1
	}

	// Convert camera position from game pixels to screen pixel offset
	screenOffsetX := c.X * ps
	screenOffsetY := c.Y * ps

	// Convert to NDC offset
	// Moving camera right (positive X) shifts world left in NDC
	// Moving camera down (positive Y) shifts world up in NDC (Y is flipped)
	shiftX := float32(-screenOffsetX * 2.0 / canvasWidth)
	shiftY := float32(screenOffsetY * 2.0 / canvasHeight)

	// Column-major 4x4 matrix: Scale(zoom) * Translate(shift)
	return [16]float32{
		z, 0, 0, 0, // column 0
		0, z, 0, 0, // column 1
		0, 0, 1, 0, // column 2
		shiftX * z, shiftY * z, 0, 1, // column 3
	}
}

// SnapZoomPixelPerfect snaps zoom to the nearest pixel-perfect value.
// Pixel-perfect means zoom is a positive integer (each game pixel = zoom * pixelScale screen pixels).
func SnapZoomPixelPerfect(zoom float64) float64 {
	snapped := math.Round(zoom)
	if snapped < 1 {
		snapped = 1
	}
	return snapped
}

// UVRect represents UV coordinates for texture sampling
type UVRect struct {
	U float64 // Left (0.0 to 1.0)
	V float64 // Top (0.0 to 1.0)
	W float64 // Width (0.0 to 1.0)
	H float64 // Height (0.0 to 1.0)
}

// Texture represents a loaded texture
type Texture interface {
	GetWidth() int
	GetHeight() int
	GetID() string
}

// WebGPUTexture implements the Texture interface
type WebGPUTexture struct {
	Width  int
	Height int
	ID     string
}

func (t *WebGPUTexture) GetWidth() int  { return t.Width }
func (t *WebGPUTexture) GetHeight() int { return t.Height }
func (t *WebGPUTexture) GetID() string  { return t.ID }

// NewWebGPUTexture creates a new WebGPUTexture with the given parameters
func NewWebGPUTexture(width, height int, id string) *WebGPUTexture {
	return &WebGPUTexture{
		Width:  width,
		Height: height,
		ID:     id,
	}
}

// Pipeline represents a WebGPU render pipeline
type Pipeline interface {
	IsValid() bool
}

// WebGPUPipeline implements the Pipeline interface
type WebGPUPipeline struct {
	Valid bool
}

func (p *WebGPUPipeline) IsValid() bool { return p.Valid }

// SpriteVertex represents a vertex for sprite rendering
type SpriteVertex struct {
	Position Vector2 // Screen position
	UV       Vector2 // Texture coordinates
}

// SpriteUniforms represents uniform data for sprite rendering
type SpriteUniforms struct {
	Transform [16]float64 // 4x4 matrix as array
	Color     [4]float64  // RGBA color
}

// DemoState represents the current demo state
type DemoState struct {
	CurrentDemo int
	DemoTime    float64
	TotalTime   float64
}

// Demo represents a single demo
type Demo struct {
	Name        string
	Description string
	Duration    float64
	Setup       func() error
	Update      func(deltaTime float64) error
	Render      func() error
	Cleanup     func() error
}

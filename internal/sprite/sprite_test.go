package sprite

import (
	"testing"

	"github.com/conor/webgpu-triangle/internal/types"
)

func TestNewSpriteSheet(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 2, 3)

	if sprite == nil {
		t.Fatal("NewSpriteSheet returned nil")
	}

	if sprite.texturePath != "test.png" {
		t.Errorf("Expected texture path 'test.png', got '%s'", sprite.texturePath)
	}

	if sprite.columns != 2 {
		t.Errorf("Expected 2 columns, got %d", sprite.columns)
	}

	if sprite.rows != 3 {
		t.Errorf("Expected 3 rows, got %d", sprite.rows)
	}

	if sprite.totalFrames != 6 {
		t.Errorf("Expected 6 total frames, got %d", sprite.totalFrames)
	}

	if !sprite.visible {
		t.Error("Expected sprite to be visible by default")
	}
}

func TestSpriteSheetGetSize(t *testing.T) {
	size := Vector2{X: 128, Y: 256}
	sprite := NewSpriteSheet("test.png", size, 2, 2)

	got := sprite.GetSize()
	if got.X != size.X || got.Y != size.Y {
		t.Errorf("Expected size %v, got %v", size, got)
	}
}

func TestSpriteSheetVisibility(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 2, 2)

	if !sprite.IsVisible() {
		t.Error("Expected sprite to be visible initially")
	}

	sprite.SetVisible(false)
	if sprite.IsVisible() {
		t.Error("Expected sprite to be invisible after SetVisible(false)")
	}

	sprite.SetVisible(true)
	if !sprite.IsVisible() {
		t.Error("Expected sprite to be visible after SetVisible(true)")
	}
}

func TestSpriteSheetFrameManagement(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 3, 2) // 6 frames

	if sprite.GetCurrentFrame() != 0 {
		t.Errorf("Expected initial frame 0, got %d", sprite.GetCurrentFrame())
	}

	if sprite.GetTotalFrames() != 6 {
		t.Errorf("Expected 6 total frames, got %d", sprite.GetTotalFrames())
	}

	sprite.SetCurrentFrame(3)
	if sprite.GetCurrentFrame() != 3 {
		t.Errorf("Expected frame 3, got %d", sprite.GetCurrentFrame())
	}

	// Test invalid frame (should be ignored)
	sprite.SetCurrentFrame(-1)
	if sprite.GetCurrentFrame() != 3 {
		t.Error("Negative frame should be ignored")
	}

	sprite.SetCurrentFrame(100)
	if sprite.GetCurrentFrame() != 3 {
		t.Error("Out-of-bounds frame should be ignored")
	}
}

func TestSpriteSheetFrameTime(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 2, 2)

	// Default frame time
	if sprite.frameTime != 0.1 {
		t.Errorf("Expected default frame time 0.1, got %f", sprite.frameTime)
	}

	sprite.SetFrameTime(0.05)
	if sprite.frameTime != 0.05 {
		t.Errorf("Expected frame time 0.05, got %f", sprite.frameTime)
	}
}

func TestSpriteSheetAnimationUpdate(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 2, 2) // 4 frames
	sprite.SetFrameTime(0.1)

	// Update less than frame time - should not advance
	sprite.Update(0.05)
	if sprite.GetCurrentFrame() != 0 {
		t.Errorf("Frame should not advance yet, got frame %d", sprite.GetCurrentFrame())
	}

	// Update to exceed frame time - should advance
	sprite.Update(0.06) // Total 0.11
	if sprite.GetCurrentFrame() != 1 {
		t.Errorf("Expected frame 1, got %d", sprite.GetCurrentFrame())
	}

	// Continue updating
	sprite.Update(0.1)
	if sprite.GetCurrentFrame() != 2 {
		t.Errorf("Expected frame 2, got %d", sprite.GetCurrentFrame())
	}

	sprite.Update(0.1)
	if sprite.GetCurrentFrame() != 3 {
		t.Errorf("Expected frame 3, got %d", sprite.GetCurrentFrame())
	}

	// Should wrap to frame 0
	sprite.Update(0.1)
	if sprite.GetCurrentFrame() != 0 {
		t.Errorf("Expected frame to wrap to 0, got %d", sprite.GetCurrentFrame())
	}
}

func TestSpriteSheetSingleFrameNoAnimation(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 1, 1) // 1 frame

	sprite.Update(1.0)
	if sprite.GetCurrentFrame() != 0 {
		t.Error("Single frame sprite should stay at frame 0")
	}
}

func TestSpriteSheetGetSpriteRenderData(t *testing.T) {
	sprite := NewSpriteSheet("texture.png", Vector2{X: 128, Y: 128}, 2, 2)
	position := types.Vector2{X: 100, Y: 200}

	renderData := sprite.GetSpriteRenderData(position)

	if renderData.TexturePath != "texture.png" {
		t.Errorf("Expected texture path 'texture.png', got '%s'", renderData.TexturePath)
	}

	if renderData.Position.X != 100 || renderData.Position.Y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", renderData.Position.X, renderData.Position.Y)
	}

	if renderData.Size.X != 128 || renderData.Size.Y != 128 {
		t.Errorf("Expected size (128, 128), got (%f, %f)", renderData.Size.X, renderData.Size.Y)
	}

	if !renderData.Visible {
		t.Error("Expected render data to be visible")
	}
}

func TestSpriteSheetUVCalculation(t *testing.T) {
	tests := []struct {
		name      string
		columns   int
		rows      int
		frame     int
		expectedU float64
		expectedV float64
		expectedW float64
		expectedH float64
	}{
		{
			name:      "2x2 Frame 0 (top-left)",
			columns:   2,
			rows:      2,
			frame:     0,
			expectedU: 0.0,
			expectedV: 0.0,
			expectedW: 0.5,
			expectedH: 0.5,
		},
		{
			name:      "2x2 Frame 1 (top-right)",
			columns:   2,
			rows:      2,
			frame:     1,
			expectedU: 0.5,
			expectedV: 0.0,
			expectedW: 0.5,
			expectedH: 0.5,
		},
		{
			name:      "2x2 Frame 2 (bottom-left)",
			columns:   2,
			rows:      2,
			frame:     2,
			expectedU: 0.0,
			expectedV: 0.5,
			expectedW: 0.5,
			expectedH: 0.5,
		},
		{
			name:      "2x2 Frame 3 (bottom-right)",
			columns:   2,
			rows:      2,
			frame:     3,
			expectedU: 0.5,
			expectedV: 0.5,
			expectedW: 0.5,
			expectedH: 0.5,
		},
		{
			name:      "3x2 Frame 4",
			columns:   3,
			rows:      2,
			frame:     4,
			expectedU: 0.333333,
			expectedV: 0.5,
			expectedW: 0.333333,
			expectedH: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, tt.columns, tt.rows)
			sprite.SetCurrentFrame(tt.frame)

			renderData := sprite.GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
			uv := renderData.UV

			if !floatEquals(uv.U, tt.expectedU, 0.0001) {
				t.Errorf("Expected U=%f, got %f", tt.expectedU, uv.U)
			}
			if !floatEquals(uv.V, tt.expectedV, 0.0001) {
				t.Errorf("Expected V=%f, got %f", tt.expectedV, uv.V)
			}
			if !floatEquals(uv.W, tt.expectedW, 0.0001) {
				t.Errorf("Expected W=%f, got %f", tt.expectedW, uv.W)
			}
			if !floatEquals(uv.H, tt.expectedH, 0.0001) {
				t.Errorf("Expected H=%f, got %f", tt.expectedH, uv.H)
			}
		})
	}
}

func TestSpriteSheetInvisibleRenderData(t *testing.T) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 2, 2)
	sprite.SetVisible(false)

	renderData := sprite.GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
	if renderData.Visible {
		t.Error("Expected render data to be invisible")
	}
}

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

func BenchmarkSpriteSheetUpdate(b *testing.B) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 4, 4)
	sprite.SetFrameTime(0.1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sprite.Update(0.016)
	}
}

func BenchmarkSpriteSheetGetRenderData(b *testing.B) {
	sprite := NewSpriteSheet("test.png", Vector2{X: 64, Y: 64}, 4, 4)
	pos := types.Vector2{X: 100, Y: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sprite.GetSpriteRenderData(pos)
	}
}

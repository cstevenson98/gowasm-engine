//go:build js

package canvas

import (
	"syscall/js"
	"testing"
	"time"

	"github.com/conor/webgpu-triangle/pkg/types"
)

// setupTestCanvas creates a test canvas element in the DOM
func setupTestCanvas(t *testing.T, canvasID string) {
	t.Helper()

	doc := js.Global().Get("document")
	canvas := doc.Call("createElement", "canvas")
	canvas.Set("id", canvasID)
	canvas.Set("width", 800)
	canvas.Set("height", 600)
	doc.Get("body").Call("appendChild", canvas)

	t.Cleanup(func() {
		if !canvas.IsNull() && !canvas.IsUndefined() {
			canvas.Call("remove")
		}
	})
}

// cleanupCanvas removes a canvas element from the DOM
func cleanupCanvas(canvasID string) {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", canvasID)
	if !canvas.IsNull() && !canvas.IsUndefined() {
		canvas.Call("remove")
	}
}

func TestNewWebGPUCanvasManager(t *testing.T) {
	manager := NewWebGPUCanvasManager()

	if manager == nil {
		t.Fatal("NewWebGPUCanvasManager returned nil")
	}

	if manager.loadedTextures == nil {
		t.Error("loadedTextures map not initialized")
	}

	if manager.initialized {
		t.Error("Expected manager to not be initialized")
	}

	if len(manager.activePipelines) != 0 {
		t.Error("Expected no active pipelines on creation")
	}
}

func TestWebGPUCanvasManager_Initialize(t *testing.T) {
	tests := []struct {
		name        string
		canvasID    string
		setupCanvas bool
		expectError bool
		expectInit  bool
		skipIfNoGPU bool
	}{
		{
			name:        "successful initialization",
			canvasID:    "test-canvas-init-1",
			setupCanvas: true,
			expectError: false,
			expectInit:  true,
			skipIfNoGPU: true,
		},
		{
			name:        "canvas not found",
			canvasID:    "nonexistent-canvas",
			setupCanvas: false,
			expectError: true,
			expectInit:  false,
			skipIfNoGPU: false,
		},
		{
			name:        "empty canvas ID",
			canvasID:    "",
			setupCanvas: false,
			expectError: true,
			expectInit:  false,
			skipIfNoGPU: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupCanvas {
				setupTestCanvas(t, tt.canvasID)
			}

			manager := NewWebGPUCanvasManager()
			err := manager.Initialize(tt.canvasID)

			// Check for WebGPU availability
			if tt.skipIfNoGPU && err != nil && err.Error() == "WebGPU not supported" {
				t.Skip("WebGPU not available in test environment")
			}

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			initialized, _ := manager.GetStatus()
			if initialized != tt.expectInit {
				t.Errorf("Expected initialized=%v, got %v", tt.expectInit, initialized)
			}

			// Cleanup if initialized
			if initialized {
				manager.Cleanup()
			}
		})
	}
}

func TestWebGPUCanvasManager_SetPipelines(t *testing.T) {
	setupTestCanvas(t, "test-canvas-pipelines")
	manager := NewWebGPUCanvasManager()

	// Test before initialization
	err := manager.SetPipelines([]types.PipelineType{types.TrianglePipeline})
	if err == nil {
		t.Error("Expected error when setting pipelines before initialization")
	}

	// Initialize
	err = manager.Initialize("test-canvas-pipelines")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	tests := []struct {
		name        string
		pipelines   []types.PipelineType
		expectError bool
	}{
		{
			name:        "single triangle pipeline",
			pipelines:   []types.PipelineType{types.TrianglePipeline},
			expectError: false,
		},
		{
			name:        "single sprite pipeline",
			pipelines:   []types.PipelineType{types.SpritePipeline},
			expectError: false,
		},
		{
			name:        "single textured pipeline",
			pipelines:   []types.PipelineType{types.TexturedPipeline},
			expectError: false,
		},
		{
			name:        "multiple pipelines",
			pipelines:   []types.PipelineType{types.TrianglePipeline, types.SpritePipeline},
			expectError: false,
		},
		{
			name:        "all pipelines",
			pipelines:   []types.PipelineType{types.TrianglePipeline, types.SpritePipeline, types.TexturedPipeline},
			expectError: false,
		},
		{
			name:        "empty pipelines",
			pipelines:   []types.PipelineType{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SetPipelines(tt.pipelines)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError && err == nil {
				if len(manager.activePipelines) != len(tt.pipelines) {
					t.Errorf("Expected %d active pipelines, got %d", len(tt.pipelines), len(manager.activePipelines))
				}
			}
		})
	}
}

func TestWebGPUCanvasManager_DrawColoredRect(t *testing.T) {
	setupTestCanvas(t, "test-canvas-draw")
	manager := NewWebGPUCanvasManager()

	// Test before initialization
	err := manager.DrawColoredRect(
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		[4]float32{1.0, 0.0, 0.0, 1.0},
	)
	if err == nil {
		t.Error("Expected error when drawing before initialization")
	}

	// Initialize
	err = manager.Initialize("test-canvas-draw")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Set sprite pipeline
	err = manager.SetPipelines([]types.PipelineType{types.SpritePipeline})
	if err != nil {
		t.Fatalf("Failed to set pipelines: %v", err)
	}

	tests := []struct {
		name      string
		position  types.Vector2
		size      types.Vector2
		color     [4]float32
		batchMode bool
	}{
		{
			name:      "immediate mode - red square",
			position:  types.Vector2{X: 100, Y: 100},
			size:      types.Vector2{X: 50, Y: 50},
			color:     [4]float32{1.0, 0.0, 0.0, 1.0},
			batchMode: false,
		},
		{
			name:      "immediate mode - green rectangle",
			position:  types.Vector2{X: 200, Y: 200},
			size:      types.Vector2{X: 100, Y: 50},
			color:     [4]float32{0.0, 1.0, 0.0, 1.0},
			batchMode: false,
		},
		{
			name:      "batch mode - blue square",
			position:  types.Vector2{X: 300, Y: 300},
			size:      types.Vector2{X: 50, Y: 50},
			color:     [4]float32{0.0, 0.0, 1.0, 1.0},
			batchMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.batchMode {
				manager.BeginBatch()
			}

			err := manager.DrawColoredRect(tt.position, tt.size, tt.color)
			if err != nil {
				t.Errorf("DrawColoredRect failed: %v", err)
			}

			if tt.batchMode {
				if len(manager.stagedVertices) == 0 {
					t.Error("Expected staged vertices in batch mode")
				}
				manager.EndBatch()
			} else {
				if manager.stagedVertexCount == 0 {
					t.Error("Expected staged vertex count > 0 in immediate mode")
				}
			}
		})
	}
}

func TestWebGPUCanvasManager_BatchRendering(t *testing.T) {
	setupTestCanvas(t, "test-canvas-batch")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-batch")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Set sprite pipeline
	err = manager.SetPipelines([]types.PipelineType{types.SpritePipeline})
	if err != nil {
		t.Fatalf("Failed to set pipelines: %v", err)
	}

	// Test BeginBatch
	err = manager.BeginBatch()
	if err != nil {
		t.Fatalf("BeginBatch failed: %v", err)
	}

	if !manager.batchMode {
		t.Error("Expected batch mode to be enabled")
	}

	if len(manager.stagedVertices) != 0 {
		t.Error("Expected empty staged vertices after BeginBatch")
	}

	// Add multiple rectangles
	for i := 0; i < 5; i++ {
		err = manager.DrawColoredRect(
			types.Vector2{X: float64(i * 100), Y: 100},
			types.Vector2{X: 50, Y: 50},
			[4]float32{1.0, 0.0, 0.0, 1.0},
		)
		if err != nil {
			t.Errorf("DrawColoredRect failed: %v", err)
		}
	}

	// Verify vertices are batched
	expectedVertices := 5 * 6 * 6 // 5 rects * 6 vertices * 6 floats per vertex
	if len(manager.stagedVertices) != expectedVertices {
		t.Errorf("Expected %d staged vertices, got %d", expectedVertices, len(manager.stagedVertices))
	}

	// Test EndBatch
	err = manager.EndBatch()
	if err != nil {
		t.Fatalf("EndBatch failed: %v", err)
	}

	if manager.batchMode {
		t.Error("Expected batch mode to be disabled after EndBatch")
	}

	if manager.stagedVertexCount == 0 {
		t.Error("Expected staged vertex count > 0 after EndBatch")
	}
}

func TestWebGPUCanvasManager_FlushBatch(t *testing.T) {
	setupTestCanvas(t, "test-canvas-flush")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-flush")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Set sprite pipeline
	err = manager.SetPipelines([]types.PipelineType{types.SpritePipeline})
	if err != nil {
		t.Fatalf("Failed to set pipelines: %v", err)
	}

	// Test FlushBatch with no vertices
	err = manager.FlushBatch()
	if err != nil {
		t.Errorf("FlushBatch failed with no vertices: %v", err)
	}

	// Test FlushBatch with vertices
	manager.BeginBatch()
	manager.DrawColoredRect(
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		[4]float32{1.0, 0.0, 0.0, 1.0},
	)

	err = manager.FlushBatch()
	if err != nil {
		t.Errorf("FlushBatch failed: %v", err)
	}

	if manager.stagedVertexCount == 0 {
		t.Error("Expected staged vertex count > 0 after flush")
	}
}

func TestWebGPUCanvasManager_CanvasToNDC(t *testing.T) {
	setupTestCanvas(t, "test-canvas-ndc")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-ndc")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	tests := []struct {
		name      string
		x, y      float64
		expectedX float32
		expectedY float32
		tolerance float32
	}{
		{
			name:      "top-left corner",
			x:         0,
			y:         0,
			expectedX: -1.0,
			expectedY: 1.0,
			tolerance: 0.01,
		},
		{
			name:      "bottom-right corner",
			x:         800,
			y:         600,
			expectedX: 1.0,
			expectedY: -1.0,
			tolerance: 0.01,
		},
		{
			name:      "center",
			x:         400,
			y:         300,
			expectedX: 0.0,
			expectedY: 0.0,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ndcX, ndcY := manager.canvasToNDC(tt.x, tt.y)

			if abs(ndcX-tt.expectedX) > tt.tolerance {
				t.Errorf("Expected NDC X=%f, got %f", tt.expectedX, ndcX)
			}

			if abs(ndcY-tt.expectedY) > tt.tolerance {
				t.Errorf("Expected NDC Y=%f, got %f", tt.expectedY, ndcY)
			}
		})
	}
}

func TestWebGPUCanvasManager_Render(t *testing.T) {
	setupTestCanvas(t, "test-canvas-render")
	manager := NewWebGPUCanvasManager()

	// Test render before initialization
	err := manager.Render()
	if err != nil {
		t.Error("Render should not error before initialization (should be no-op)")
	}

	// Initialize
	err = manager.Initialize("test-canvas-render")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Set triangle pipeline
	err = manager.SetPipelines([]types.PipelineType{types.TrianglePipeline})
	if err != nil {
		t.Fatalf("Failed to set pipelines: %v", err)
	}

	// Test render after initialization
	err = manager.Render()
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}

	// Test multiple renders
	for i := 0; i < 3; i++ {
		err = manager.Render()
		if err != nil {
			t.Errorf("Render %d failed: %v", i, err)
		}
	}
}

func TestWebGPUCanvasManager_LoadTexture(t *testing.T) {
	setupTestCanvas(t, "test-canvas-texture")
	manager := NewWebGPUCanvasManager()

	// Test before initialization
	err := manager.LoadTexture("test.png")
	if err == nil {
		t.Error("Expected error when loading texture before initialization")
	}

	// Initialize
	err = manager.Initialize("test-canvas-texture")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Note: Actual texture loading requires a valid image URL and will be async
	// This test just verifies the method can be called without panic
	err = manager.LoadTexture("test-texture.png")
	if err != nil {
		t.Logf("LoadTexture returned error (expected in test env): %v", err)
	}

	// Give a small delay for async operations
	time.Sleep(100 * time.Millisecond)
}

func TestWebGPUCanvasManager_DrawTexturedRect(t *testing.T) {
	setupTestCanvas(t, "test-canvas-textured")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-textured")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Set textured pipeline
	err = manager.SetPipelines([]types.PipelineType{types.TexturedPipeline})
	if err != nil {
		t.Fatalf("Failed to set pipelines: %v", err)
	}

	// Test drawing without loaded texture (should error)
	err = manager.DrawTexturedRect(
		"nonexistent.png",
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		types.UVRect{U: 0, V: 0, W: 1, H: 1},
	)
	if err == nil {
		t.Error("Expected error when drawing textured rect with unloaded texture")
	}
}

func TestWebGPUCanvasManager_GetStatus(t *testing.T) {
	manager := NewWebGPUCanvasManager()

	// Test initial status
	initialized, message := manager.GetStatus()
	if initialized {
		t.Error("Expected initial status to be uninitialized")
	}
	if message != "" {
		t.Errorf("Expected empty message, got: %s", message)
	}

	// Test SetStatus
	manager.SetStatus(true, "Test message")
	initialized, message = manager.GetStatus()
	if !initialized {
		t.Error("Expected initialized status after SetStatus")
	}
	if message != "Test message" {
		t.Errorf("Expected 'Test message', got: %s", message)
	}
}

func TestWebGPUCanvasManager_Cleanup(t *testing.T) {
	setupTestCanvas(t, "test-canvas-cleanup")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-cleanup")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Verify initialized
	initialized, _ := manager.GetStatus()
	if !initialized {
		t.Fatal("Expected manager to be initialized before cleanup")
	}

	// Cleanup
	err = manager.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Verify status after cleanup
	initialized, _ = manager.GetStatus()
	if initialized {
		t.Error("Expected manager to be uninitialized after cleanup")
	}
}

func TestWebGPUCanvasManager_StubMethods(t *testing.T) {
	setupTestCanvas(t, "test-canvas-stubs")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-stubs")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Test ClearCanvas
	err = manager.ClearCanvas()
	if err != nil {
		t.Errorf("ClearCanvas failed: %v", err)
	}

	// Test GetSpritePipeline
	pipeline := manager.GetSpritePipeline()
	if pipeline == nil {
		t.Error("GetSpritePipeline returned nil")
	}

	// Test GetBackgroundPipeline
	bgPipeline := manager.GetBackgroundPipeline()
	if bgPipeline == nil {
		t.Error("GetBackgroundPipeline returned nil")
	}

	// Test DrawTexture (stub that falls back to colored rect)
	texture := types.NewWebGPUTexture(100, 100, "test-texture")
	err = manager.DrawTexture(
		texture,
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		types.UVRect{U: 0, V: 0, W: 1, H: 1},
	)
	if err != nil {
		t.Errorf("DrawTexture failed: %v", err)
	}

	// Test DrawTextureRotated (stub - should not panic)
	err = manager.DrawTextureRotated(
		texture,
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		types.UVRect{U: 0, V: 0, W: 1, H: 1},
		45.0,
	)
	if err != nil {
		t.Errorf("DrawTextureRotated failed: %v", err)
	}

	// Test DrawTextureScaled (stub - should not panic)
	err = manager.DrawTextureScaled(
		texture,
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		types.UVRect{U: 0, V: 0, W: 1, H: 1},
		types.Vector2{X: 2.0, Y: 2.0},
	)
	if err != nil {
		t.Errorf("DrawTextureScaled failed: %v", err)
	}
}

func TestWebGPUCanvasManager_GenerateQuadVertices(t *testing.T) {
	setupTestCanvas(t, "test-canvas-vertices")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-vertices")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	pos := types.Vector2{X: 100, Y: 100}
	size := types.Vector2{X: 50, Y: 50}
	color := [4]float32{1.0, 0.0, 0.0, 1.0}

	vertices := manager.generateQuadVertices(pos, size, color)

	// Should generate 6 vertices (2 triangles) * 6 floats per vertex = 36 floats
	expectedLength := 36
	if len(vertices) != expectedLength {
		t.Errorf("Expected %d vertices, got %d", expectedLength, len(vertices))
	}

	// Verify color values are present
	for i := 0; i < 6; i++ {
		colorOffset := i * 6
		r := vertices[colorOffset+2]
		g := vertices[colorOffset+3]
		b := vertices[colorOffset+4]
		a := vertices[colorOffset+5]

		if r != color[0] || g != color[1] || b != color[2] || a != color[3] {
			t.Errorf("Color mismatch at vertex %d: expected [%f,%f,%f,%f], got [%f,%f,%f,%f]",
				i, color[0], color[1], color[2], color[3], r, g, b, a)
		}
	}
}

func TestWebGPUCanvasManager_GenerateTexturedQuadVertices(t *testing.T) {
	setupTestCanvas(t, "test-canvas-tex-vertices")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-tex-vertices")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	pos := types.Vector2{X: 100, Y: 100}
	size := types.Vector2{X: 50, Y: 50}
	uv := types.UVRect{U: 0.0, V: 0.0, W: 1.0, H: 1.0}

	vertices := manager.generateTexturedQuadVertices(pos, size, uv)

	// Should generate 6 vertices (2 triangles) * 4 floats per vertex = 24 floats
	expectedLength := 24
	if len(vertices) != expectedLength {
		t.Errorf("Expected %d vertices, got %d", expectedLength, len(vertices))
	}

	// Verify UV coordinates are present
	// Each vertex has 4 floats: [x, y, u, v]
	uvCoords := [][2]float32{
		{0.0, 0.0}, {1.0, 0.0}, {0.0, 1.0}, // Triangle 1
		{1.0, 0.0}, {1.0, 1.0}, {0.0, 1.0}, // Triangle 2
	}

	for i := 0; i < 6; i++ {
		uvOffset := i * 4
		u := vertices[uvOffset+2]
		v := vertices[uvOffset+3]

		expectedU := uvCoords[i][0]
		expectedV := uvCoords[i][1]

		if abs(u-expectedU) > 0.01 || abs(v-expectedV) > 0.01 {
			t.Errorf("UV mismatch at vertex %d: expected [%f,%f], got [%f,%f]",
				i, expectedU, expectedV, u, v)
		}
	}
}

func TestWebGPUCanvasManager_PipelineSwitching(t *testing.T) {
	setupTestCanvas(t, "test-canvas-pipeline-switch")
	manager := NewWebGPUCanvasManager()

	// Initialize
	err := manager.Initialize("test-canvas-pipeline-switch")
	if err != nil {
		if err.Error() == "WebGPU not supported" {
			t.Skip("WebGPU not available in test environment")
		}
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Cleanup()

	// Test switching between different pipelines
	pipelines := [][]types.PipelineType{
		{types.TrianglePipeline},
		{types.SpritePipeline},
		{types.TexturedPipeline},
		{types.TrianglePipeline, types.SpritePipeline},
		{types.SpritePipeline, types.TexturedPipeline},
	}

	for i, pipelineSet := range pipelines {
		err := manager.SetPipelines(pipelineSet)
		if err != nil {
			t.Errorf("Failed to set pipeline set %d: %v", i, err)
		}

		// Verify staged vertices are cleared on pipeline switch
		if manager.stagedVertexCount != 0 {
			t.Errorf("Expected staged vertex count to be 0 after pipeline switch %d", i)
		}

		if len(manager.stagedVertices) != 0 {
			t.Errorf("Expected staged vertices to be cleared after pipeline switch %d", i)
		}
	}
}

func TestFloat32SliceToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected int // expected byte length
	}{
		{
			name:     "empty slice",
			input:    []float32{},
			expected: 0,
		},
		{
			name:     "single float",
			input:    []float32{1.0},
			expected: 4,
		},
		{
			name:     "multiple floats",
			input:    []float32{1.0, 2.0, 3.0, 4.0},
			expected: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := float32SliceToBytes(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected byte length %d, got %d", tt.expected, len(result))
			}
		})
	}
}

// Helper function for floating point comparison
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

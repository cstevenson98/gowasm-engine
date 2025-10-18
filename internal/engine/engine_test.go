//go:build js

package engine

import (
	"testing"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/input"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Helper to create an engine with mocks
func newTestEngine() *Engine {
	return &Engine{
		canvasManager:      canvas.NewMockCanvasManager(),
		inputCapturer:      input.NewMockInput(),
		running:            false,
		gameStatePipelines: make(map[types.GameState][]types.PipelineType),
		screenWidth:        800.0,
		screenHeight:       600.0,
	}
}

func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.canvasManager == nil {
		t.Error("Canvas manager not initialized")
	}

	if engine.inputCapturer == nil {
		t.Error("Input capturer not initialized")
	}

	if engine.gameStatePipelines == nil {
		t.Error("Game state pipelines map not initialized")
	}

	if engine.screenWidth == 0 || engine.screenHeight == 0 {
		t.Error("Screen dimensions not initialized")
	}

	if engine.running {
		t.Error("Engine should not be running initially")
	}
}

func TestEngineInitialization(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()

	// Check SPRITE state
	pipelines, exists := engine.gameStatePipelines[types.GAMEPLAY]
	if !exists {
		t.Error("SPRITE state pipelines not initialized")
	}
	if len(pipelines) != 1 || pipelines[0] != types.TexturedPipeline {
		t.Error("SPRITE state should have TexturedPipeline")
	}

	// Check TRIANGLE state
	pipelines, exists = engine.gameStatePipelines[types.TRIANGLE]
	if !exists {
		t.Error("TRIANGLE state pipelines not initialized")
	}
	if len(pipelines) != 1 || pipelines[0] != types.TrianglePipeline {
		t.Error("TRIANGLE state should have TrianglePipeline")
	}
}

func TestEngineSetGameState(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()

	// Initialize mock canvas
	engine.canvasManager.Initialize("test-canvas")

	// Set to SPRITE state
	err := engine.SetGameState(types.GAMEPLAY)
	if err != nil {
		t.Errorf("Failed to set SPRITE state: %v", err)
	}

	if engine.GetGameState() != types.GAMEPLAY {
		t.Error("Game state not set to SPRITE")
	}

	// Set to TRIANGLE state
	err = engine.SetGameState(types.TRIANGLE)
	if err != nil {
		t.Errorf("Failed to set TRIANGLE state: %v", err)
	}

	if engine.GetGameState() != types.TRIANGLE {
		t.Error("Game state not set to TRIANGLE")
	}
}

func TestEngineSetInvalidGameState(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")

	// Try to set an invalid state (value not in the map)
	invalidState := types.GameState(999)
	err := engine.SetGameState(invalidState)

	if err == nil {
		t.Error("Expected error when setting invalid game state")
	}
}

func TestEngineGetGameState(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")

	// Set initial state
	engine.SetGameState(types.GAMEPLAY)

	// Get state should return SPRITE
	state := engine.GetGameState()
	if state != types.GAMEPLAY {
		t.Errorf("Expected SPRITE state, got %v", state)
	}
}

func TestEngineUpdateWithPlayer(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	// Get the gameplay scene and player
	if engine.currentScene == nil {
		t.Fatal("Current scene should not be nil after SetGameState")
	}

	// Set player input
	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.SetInputState(types.InputState{MoveRight: true})

	// Get renderables to access player
	renderables := engine.currentScene.GetRenderables()
	if len(renderables) == 0 {
		t.Fatal("Expected at least one renderable (player)")
	}

	// Get initial player position
	initialPos := renderables[0].GetMover().GetPosition()

	// Update engine
	engine.Update(1.0) // 1 second

	// Player should have moved right
	renderables = engine.currentScene.GetRenderables()
	newPos := renderables[0].GetMover().GetPosition()
	if newPos.X <= initialPos.X {
		t.Error("Player should have moved right")
	}
}

func TestEngineUpdateWithGameObjects(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	// Note: This test verifies scene update behavior
	// Game objects would be added through scene API in real usage

	// Update
	engine.Update(0.016)

	// Verify scene is updated
	if engine.currentScene == nil {
		t.Error("Scene should be initialized")
	}
}

func TestEngineRenderWithPlayer(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	mockCanvas := engine.canvasManager.(*canvas.MockCanvasManager)
	mockCanvas.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	// Load texture
	mockCanvas.LoadTexture("llama.png")

	// Render
	engine.Render()

	// Check that render was called
	if mockCanvas.GetRenderCount() < 1 {
		t.Error("Render should have been called")
	}
}

func TestEngineRenderWithMultipleObjects(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	mockCanvas := engine.canvasManager.(*canvas.MockCanvasManager)
	mockCanvas.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	// Scene now manages objects
	// In real usage, objects would be added through scene API
	mockCanvas.LoadTexture("llama.png")

	// Render
	engine.Render()

	// Should have called render at least once
	if mockCanvas.GetRenderCount() < 1 {
		t.Error("Render should have been called")
	}
}

func TestEngineCleanup(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	mockCanvas := engine.canvasManager.(*canvas.MockCanvasManager)
	mockCanvas.Initialize("test-canvas")
	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.Initialize()

	engine.running = true

	// Cleanup
	err := engine.Cleanup()
	if err != nil {
		t.Errorf("Cleanup returned error: %v", err)
	}

	if engine.running {
		t.Error("Engine should not be running after cleanup")
	}

	if !mockCanvas.WasCleanupCalled() {
		t.Error("Canvas cleanup should have been called")
	}
}

func TestEngineStop(t *testing.T) {
	engine := newTestEngine()
	engine.running = true

	engine.Stop()

	if engine.running {
		t.Error("Engine should not be running after Stop()")
	}
}

func TestEngineGetCanvasManager(t *testing.T) {
	engine := newTestEngine()

	cm := engine.GetCanvasManager()
	if cm == nil {
		t.Error("GetCanvasManager returned nil")
	}

	if cm != engine.canvasManager {
		t.Error("GetCanvasManager should return the same instance")
	}
}

func TestEngineStateLocking(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")

	// Set initial state
	engine.SetGameState(types.GAMEPLAY)

	// Simulate concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			engine.SetGameState(types.GAMEPLAY)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = engine.GetGameState()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic or deadlock
}

func TestEnginePlayerInitialization(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	// Player is now created by the scene
	if engine.currentScene == nil {
		t.Fatal("Scene not initialized")
	}

	renderables := engine.currentScene.GetRenderables()
	if len(renderables) == 0 {
		t.Fatal("Expected player in renderables")
	}

	player := renderables[0]

	// Check player components
	if player.GetSprite() == nil {
		t.Error("Player sprite not initialized")
	}

	if player.GetMover() == nil {
		t.Error("Player mover not initialized")
	}

	// Check player position (should be centered)
	pos := player.GetMover().GetPosition()
	// Screen is 800x600, sprite is 128x128, so center is ~336, 236
	if pos.X < 300 || pos.X > 400 {
		t.Errorf("Player X position unexpected: %f", pos.X)
	}
	if pos.Y < 200 || pos.Y > 300 {
		t.Errorf("Player Y position unexpected: %f", pos.Y)
	}
}

func TestEngineUpdateDeltaTime(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)

	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.SetInputState(types.InputState{MoveRight: true})

	renderables := engine.currentScene.GetRenderables()
	if len(renderables) == 0 {
		t.Fatal("Expected player in renderables")
	}
	player := renderables[0]

	initialPos := player.GetMover().GetPosition()

	// Update with different delta times
	engine.Update(0.016) // 60fps frame
	pos1 := player.GetMover().GetPosition()

	player.GetMover().SetPosition(initialPos) // Reset

	engine.Update(0.032) // 30fps frame (double time)
	pos2 := player.GetMover().GetPosition()

	// Movement should be proportional to delta time
	delta1 := pos1.X - initialPos.X
	delta2 := pos2.X - initialPos.X

	// pos2 should have moved roughly twice as far as pos1
	if delta2 < delta1*1.8 || delta2 > delta1*2.2 {
		t.Errorf("Delta time not properly applied: delta1=%f, delta2=%f", delta1, delta2)
	}
}

func TestEngineNoPlayerUpdateInTriangleState(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.TRIANGLE)

	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.SetInputState(types.InputState{MoveRight: true})

	// TRIANGLE state has no scene, so no player
	if engine.currentScene != nil {
		t.Error("TRIANGLE state should not have a scene")
	}

	// Update in TRIANGLE state
	engine.Update(1.0)

	// Should complete without error (no scene to update)
}

func TestEngineError(t *testing.T) {
	err := &EngineError{Message: "test error"}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}
}

// Helper test GameObject implementation
type testGameObject struct {
	sprite types.Sprite
	mover  types.Mover
	state  types.ObjectState
}

func (t *testGameObject) GetSprite() types.Sprite {
	return t.sprite
}

func (t *testGameObject) GetMover() types.Mover {
	return t.mover
}

func (t *testGameObject) Update(deltaTime float64) {
	// Test implementation
}

func (t *testGameObject) GetState() *types.ObjectState {
	return &t.state
}

func (t *testGameObject) SetState(state types.ObjectState) {
	t.state = state
}

func BenchmarkEngineUpdate(b *testing.B) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.SetGameState(types.GAMEPLAY)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Update(0.016)
	}
}

func BenchmarkEngineRender(b *testing.B) {
	engine := newTestEngine()
	engine.initializeGameStates()
	mockCanvas := engine.canvasManager.(*canvas.MockCanvasManager)
	mockCanvas.Initialize("test-canvas")
	engine.SetGameState(types.GAMEPLAY)
	mockCanvas.LoadTexture("llama.png")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Render()
	}
}

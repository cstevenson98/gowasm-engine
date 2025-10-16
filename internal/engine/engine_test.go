//go:build js

package engine

import (
	"testing"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/input"
	"github.com/conor/webgpu-triangle/internal/mover"
	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Helper to create an engine with mocks
func newTestEngine() *Engine {
	return &Engine{
		canvasManager:        canvas.NewMockCanvasManager(),
		inputCapturer:        input.NewMockInput(),
		running:              false,
		gameStatePipelines:   make(map[types.GameState][]types.PipelineType),
		gameStateGameObjects: make(map[types.GameState][]types.GameObject),
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

	if engine.gameStateGameObjects == nil {
		t.Error("Game state game objects map not initialized")
	}

	if engine.running {
		t.Error("Engine should not be running initially")
	}
}

func TestEngineInitialization(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()

	// Check SPRITE state
	pipelines, exists := engine.gameStatePipelines[types.SPRITE]
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

	// Check player was created
	if engine.player == nil {
		t.Error("Player not created during initialization")
	}
}

func TestEngineSetGameState(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()

	// Initialize mock canvas
	engine.canvasManager.Initialize("test-canvas")

	// Set to SPRITE state
	err := engine.SetGameState(types.SPRITE)
	if err != nil {
		t.Errorf("Failed to set SPRITE state: %v", err)
	}

	if engine.GetGameState() != types.SPRITE {
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
	engine.SetGameState(types.SPRITE)

	// Get state should return SPRITE
	state := engine.GetGameState()
	if state != types.SPRITE {
		t.Errorf("Expected SPRITE state, got %v", state)
	}
}

func TestEngineUpdateWithPlayer(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.canvasManager.Initialize("test-canvas")
	engine.SetGameState(types.SPRITE)

	// Set player input
	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.SetInputState(types.InputState{MoveRight: true})

	// Get initial player position
	initialPos := engine.player.GetMover().GetPosition()

	// Update engine
	engine.Update(1.0) // 1 second

	// Player should have moved right
	newPos := engine.player.GetMover().GetPosition()
	if newPos.X <= initialPos.X {
		t.Error("Player should have moved right")
	}
}

func TestEngineUpdateWithGameObjects(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	engine.SetGameState(types.SPRITE)

	// Create a test game object with mock components
	mockSprite := sprite.NewMockSprite("test.png", types.Vector2{X: 64, Y: 64})
	mockMover := mover.NewMockMover(
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 10, Y: 0},
	)

	testObject := &testGameObject{
		sprite: mockSprite,
		mover:  mockMover,
	}

	// Add to game state
	engine.gameStateGameObjects[types.SPRITE] = []types.GameObject{testObject}

	// Update
	engine.Update(0.016)

	// Verify updates were called
	if !mockSprite.WasUpdateCalled() {
		t.Error("Sprite Update should have been called")
	}
	if !mockMover.WasUpdateCalled() {
		t.Error("Mover Update should have been called")
	}
}

func TestEngineRenderWithPlayer(t *testing.T) {
	engine := newTestEngine()
	engine.initializeGameStates()
	mockCanvas := engine.canvasManager.(*canvas.MockCanvasManager)
	mockCanvas.Initialize("test-canvas")
	engine.SetGameState(types.SPRITE)

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
	engine.SetGameState(types.SPRITE)

	// Create multiple test objects
	objects := make([]types.GameObject, 3)
	for i := 0; i < 3; i++ {
		mockSprite := sprite.NewMockSprite("test.png", types.Vector2{X: 64, Y: 64})
		mockMover := mover.NewMockMover(
			types.Vector2{X: float64(i * 100), Y: 100},
			types.Vector2{X: 0, Y: 0},
		)
		objects[i] = &testGameObject{
			sprite: mockSprite,
			mover:  mockMover,
		}
	}

	engine.gameStateGameObjects[types.SPRITE] = objects
	mockCanvas.LoadTexture("test.png")

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
	engine.SetGameState(types.SPRITE)

	// Simulate concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			engine.SetGameState(types.SPRITE)
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

	if engine.player == nil {
		t.Fatal("Player not initialized")
	}

	// Check player components
	if engine.player.GetSprite() == nil {
		t.Error("Player sprite not initialized")
	}

	if engine.player.GetMover() == nil {
		t.Error("Player mover not initialized")
	}

	// Check player position (should be centered)
	pos := engine.player.GetMover().GetPosition()
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
	engine.SetGameState(types.SPRITE)

	mockInput := engine.inputCapturer.(*input.MockInput)
	mockInput.SetInputState(types.InputState{MoveRight: true})

	initialPos := engine.player.GetMover().GetPosition()

	// Update with different delta times
	engine.Update(0.016) // 60fps frame
	pos1 := engine.player.GetMover().GetPosition()

	engine.player.GetMover().SetPosition(initialPos) // Reset

	engine.Update(0.032) // 30fps frame (double time)
	pos2 := engine.player.GetMover().GetPosition()

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

	initialPos := engine.player.GetMover().GetPosition()

	// Update in TRIANGLE state
	engine.Update(1.0)

	// Player should NOT have moved (no player update in TRIANGLE state)
	newPos := engine.player.GetMover().GetPosition()
	if newPos.X != initialPos.X || newPos.Y != initialPos.Y {
		t.Error("Player should not move in TRIANGLE state")
	}
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
	engine.SetGameState(types.SPRITE)

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
	engine.SetGameState(types.SPRITE)
	mockCanvas.LoadTexture("llama.png")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Render()
	}
}

//go:build js

package scene

import (
	"testing"

	"github.com/conor/webgpu-triangle/internal/input"
	"github.com/conor/webgpu-triangle/internal/types"
)

func TestSceneLayerString(t *testing.T) {
	tests := []struct {
		layer    SceneLayer
		expected string
	}{
		{BACKGROUND, "BACKGROUND"},
		{ENTITIES, "ENTITIES"},
		{UI, "UI"},
		{SceneLayer(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.layer.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.layer.String())
		}
	}
}

func TestNewGameplayScene(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)

	if scene == nil {
		t.Fatal("NewGameplayScene returned nil")
	}

	if scene.GetName() != "Gameplay" {
		t.Errorf("Expected name 'Gameplay', got '%s'", scene.GetName())
	}

	if scene.screenWidth != 800 || scene.screenHeight != 600 {
		t.Error("Screen dimensions not set correctly")
	}
}

func TestGameplaySceneInitialize(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)

	err := scene.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check that layers are initialized
	if scene.layers == nil {
		t.Error("Layers not initialized")
	}

	// Check that player was created
	renderables := scene.GetRenderables()
	if len(renderables) == 0 {
		t.Error("Player should be in renderables after initialization")
	}
}

func TestGameplaySceneGetRenderables(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)
	scene.Initialize()

	renderables := scene.GetRenderables()

	// Should have at least the player
	if len(renderables) < 1 {
		t.Error("Expected at least player in renderables")
	}

	// Player should be first (in ENTITIES layer)
	player := renderables[0]
	if player.GetSprite() == nil {
		t.Error("Player sprite should not be nil")
	}
	if player.GetMover() == nil {
		t.Error("Player mover should not be nil")
	}
}

func TestGameplaySceneUpdate(t *testing.T) {
	mockInput := input.NewMockInput()
	mockInput.Initialize()
	scene := NewGameplayScene(800, 600, mockInput)
	scene.Initialize()

	// Set input to move right
	mockInput.SetInputState(types.InputState{MoveRight: true})

	renderables := scene.GetRenderables()
	initialPos := renderables[0].GetMover().GetPosition()

	// Update scene
	scene.Update(0.1) // 100ms

	// Player should have moved
	newPos := renderables[0].GetMover().GetPosition()
	if newPos.X <= initialPos.X {
		t.Error("Player should have moved right")
	}
}

func TestGameplaySceneCleanup(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)
	scene.Initialize()

	// Verify player exists
	if scene.player == nil {
		t.Fatal("Player should exist before cleanup")
	}

	// Cleanup
	scene.Cleanup()

	// Player should be nil after cleanup
	if scene.player != nil {
		t.Error("Player should be nil after cleanup")
	}

	// Layers should be cleared
	if len(scene.layers) != 0 {
		t.Error("Layers should be cleared after cleanup")
	}
}

func TestGameplaySceneGetPlayer(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)
	scene.Initialize()

	player := scene.GetPlayer()
	if player == nil {
		t.Error("GetPlayer should return player")
	}

	if player.GetSprite() == nil {
		t.Error("Player should have sprite")
	}
}

func TestGameplayScenePlayerCentered(t *testing.T) {
	mockInput := input.NewMockInput()
	scene := NewGameplayScene(800, 600, mockInput)
	scene.Initialize()

	player := scene.GetPlayer()
	pos := player.GetMover().GetPosition()

	// Player should be roughly centered
	// Screen 800x600, player 128x128, so center is (336, 236)
	if pos.X < 300 || pos.X > 400 {
		t.Errorf("Player X position not centered: %f", pos.X)
	}
	if pos.Y < 200 || pos.Y > 300 {
		t.Errorf("Player Y position not centered: %f", pos.Y)
	}
}

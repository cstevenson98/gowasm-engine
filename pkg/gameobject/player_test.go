//go:build js

package gameobject

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/mover"
	"github.com/cstevenson98/gowasm-engine/pkg/sprite"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func TestNewPlayer(t *testing.T) {
	position := types.Vector2{X: 100, Y: 200}
	size := types.Vector2{X: 64, Y: 64}
	speed := 150.0

	player := NewPlayer(position, size, speed)

	if player == nil {
		t.Fatal("NewPlayer returned nil")
	}

	if player.GetSprite() == nil {
		t.Error("Player sprite is nil")
	}

	if player.GetMover() == nil {
		t.Error("Player mover is nil")
	}

	if player.moveSpeed != speed {
		t.Errorf("Expected move speed %f, got %f", speed, player.moveSpeed)
	}

	state := player.GetState()
	if state == nil {
		t.Fatal("Player state is nil")
	}

	if state.ID == "" {
		t.Error("Player ID should not be empty")
	}

	if !state.Visible {
		t.Error("Player should be visible by default")
	}
}

func TestPlayerHandleInputNoMovement(t *testing.T) {
	player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)

	inputState := types.InputState{
		MoveUp:    false,
		MoveDown:  false,
		MoveLeft:  false,
		MoveRight: false,
	}

	player.HandleInput(inputState)

	vel := player.GetMover().GetVelocity()
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("Expected zero velocity, got (%f, %f)", vel.X, vel.Y)
	}
}

func TestPlayerHandleInputSingleDirection(t *testing.T) {
	tests := []struct {
		name       string
		inputState types.InputState
		expectedX  float64
		expectedY  float64
	}{
		{
			name:       "Move Up",
			inputState: types.InputState{MoveUp: true},
			expectedX:  0,
			expectedY:  -100,
		},
		{
			name:       "Move Down",
			inputState: types.InputState{MoveDown: true},
			expectedX:  0,
			expectedY:  100,
		},
		{
			name:       "Move Left",
			inputState: types.InputState{MoveLeft: true},
			expectedX:  -100,
			expectedY:  0,
		},
		{
			name:       "Move Right",
			inputState: types.InputState{MoveRight: true},
			expectedX:  100,
			expectedY:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)
			player.HandleInput(tt.inputState)

			vel := player.GetMover().GetVelocity()
			if !floatEquals(vel.X, tt.expectedX, 0.001) || !floatEquals(vel.Y, tt.expectedY, 0.001) {
				t.Errorf("Expected velocity (%f, %f), got (%f, %f)", tt.expectedX, tt.expectedY, vel.X, vel.Y)
			}
		})
	}
}

func TestPlayerHandleInputDiagonalMovement(t *testing.T) {
	player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)

	// Diagonal movement should be normalized
	inputState := types.InputState{
		MoveUp:    true,
		MoveRight: true,
	}

	player.HandleInput(inputState)
	vel := player.GetMover().GetVelocity()

	// Each component should be ~70.71 (100 * 1/sqrt(2))
	expectedComponent := 100.0 * 0.7071

	if !floatEquals(vel.X, expectedComponent, 0.01) {
		t.Errorf("Expected X velocity ~%f, got %f", expectedComponent, vel.X)
	}
	if !floatEquals(vel.Y, -expectedComponent, 0.01) {
		t.Errorf("Expected Y velocity ~%f, got %f", -expectedComponent, vel.Y)
	}

	// Total speed should still be approximately 100
	totalSpeed := vel.X*vel.X + vel.Y*vel.Y
	expectedSpeed := 100.0 * 100.0
	if !floatEquals(totalSpeed, expectedSpeed, 1.0) {
		t.Errorf("Diagonal speed should be preserved: expected %f, got %f", expectedSpeed, totalSpeed)
	}
}

func TestPlayerHandleInputOppositeDirections(t *testing.T) {
	player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)

	// Pressing opposite directions should cancel out
	inputState := types.InputState{
		MoveUp:    true,
		MoveDown:  true,
		MoveLeft:  true,
		MoveRight: true,
	}

	player.HandleInput(inputState)
	vel := player.GetMover().GetVelocity()

	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("Opposite directions should cancel, got velocity (%f, %f)", vel.X, vel.Y)
	}
}

func TestPlayerHandleInputDifferentSpeeds(t *testing.T) {
	speeds := []float64{50.0, 100.0, 200.0, 500.0}

	for _, speed := range speeds {
		player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, speed)
		inputState := types.InputState{MoveRight: true}

		player.HandleInput(inputState)
		vel := player.GetMover().GetVelocity()

		if !floatEquals(vel.X, speed, 0.001) {
			t.Errorf("Expected velocity %f, got %f", speed, vel.X)
		}
	}
}

func TestPlayerUpdate(t *testing.T) {
	player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)

	// Update should not crash
	player.Update(0.016)
	player.Update(1.0)
	player.Update(0.0)
}

func TestPlayerGetSetState(t *testing.T) {
	player := NewPlayer(types.Vector2{X: 100, Y: 200}, types.Vector2{X: 64, Y: 64}, 100.0)

	// Copy the original state value (not pointer)
	originalState := *player.GetState()
	originalID := originalState.ID

	newState := types.ObjectState{
		ID:       "test-id",
		Position: types.Vector2{X: 300, Y: 400},
		Visible:  false,
		Frame:    5,
	}

	player.SetState(newState)

	retrievedState := player.GetState()
	if retrievedState.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", retrievedState.ID)
	}

	// Verify original state copy wasn't modified
	if originalState.ID != originalID {
		t.Error("Original state copy was modified")
	}
}

func TestPlayerIntegration(t *testing.T) {
	// Integration test: input -> velocity -> position
	player := NewPlayer(types.Vector2{X: 100, Y: 100}, types.Vector2{X: 64, Y: 64}, 100.0)

	// Set input to move right
	inputState := types.InputState{MoveRight: true}
	player.HandleInput(inputState)

	// Update mover to apply velocity
	player.GetMover().Update(1.0) // 1 second

	// Position should have moved 100 pixels right
	pos := player.GetMover().GetPosition()
	if !floatEquals(pos.X, 200, 0.001) {
		t.Errorf("Expected X position 200, got %f", pos.X)
	}
	if !floatEquals(pos.Y, 100, 0.001) {
		t.Errorf("Expected Y position 100, got %f", pos.Y)
	}
}

func TestPlayerWithMockComponents(t *testing.T) {
	// Test player with mock sprite and mover
	mockSprite := sprite.NewMockSprite("test.png", types.Vector2{X: 64, Y: 64})
	mockMover := mover.NewMockMover(
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 0, Y: 0},
	)

	// Create player manually with mocks
	player := &Player{
		sprite:    mockSprite,
		mover:     mockMover,
		moveSpeed: 100.0,
		state: types.ObjectState{
			ID:       "test",
			Position: types.Vector2{X: 100, Y: 100},
			Visible:  true,
		},
	}

	// Test that we can interact with mocked components
	if player.GetSprite() == nil {
		t.Error("Mock sprite should not be nil")
	}

	if player.GetMover() == nil {
		t.Error("Mock mover should not be nil")
	}

	// Test handle input with mock
	inputState := types.InputState{MoveUp: true}
	player.HandleInput(inputState)

	vel := player.GetMover().GetVelocity()
	if vel.Y != -100 {
		t.Errorf("Expected Y velocity -100, got %f", vel.Y)
	}
}

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

func BenchmarkPlayerHandleInput(b *testing.B) {
	player := NewPlayer(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 64, Y: 64}, 100.0)
	inputState := types.InputState{MoveRight: true, MoveUp: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.HandleInput(inputState)
	}
}

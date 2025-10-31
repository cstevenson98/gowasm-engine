package mover

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func TestNewBasicMover(t *testing.T) {
	position := types.Vector2{X: 100, Y: 200}
	velocity := types.Vector2{X: 50, Y: -30}

	mover := NewBasicMover(position, velocity, 64, 64)

	if mover == nil {
		t.Fatal("NewBasicMover returned nil")
	}

	if mover.GetPosition() != position {
		t.Errorf("Expected position %v, got %v", position, mover.GetPosition())
	}

	if mover.GetVelocity() != velocity {
		t.Errorf("Expected velocity %v, got %v", velocity, mover.GetVelocity())
	}
}

func TestBasicMoverGetSetPosition(t *testing.T) {
	mover := NewBasicMover(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 0, Y: 0}, 64, 64)

	newPos := types.Vector2{X: 123.456, Y: 789.012}
	mover.SetPosition(newPos)

	got := mover.GetPosition()
	if got.X != newPos.X || got.Y != newPos.Y {
		t.Errorf("Expected position %v, got %v", newPos, got)
	}
}

func TestBasicMoverGetSetVelocity(t *testing.T) {
	mover := NewBasicMover(types.Vector2{X: 0, Y: 0}, types.Vector2{X: 0, Y: 0}, 64, 64)

	newVel := types.Vector2{X: 100, Y: -50}
	mover.SetVelocity(newVel)

	got := mover.GetVelocity()
	if got.X != newVel.X || got.Y != newVel.Y {
		t.Errorf("Expected velocity %v, got %v", newVel, got)
	}
}

func TestBasicMoverUpdate(t *testing.T) {
	tests := []struct {
		name        string
		startPos    types.Vector2
		velocity    types.Vector2
		deltaTime   float64
		expectedPos types.Vector2
	}{
		{
			name:        "Stationary",
			startPos:    types.Vector2{X: 100, Y: 100},
			velocity:    types.Vector2{X: 0, Y: 0},
			deltaTime:   1.0,
			expectedPos: types.Vector2{X: 100, Y: 100},
		},
		{
			name:        "Move Right",
			startPos:    types.Vector2{X: 0, Y: 0},
			velocity:    types.Vector2{X: 100, Y: 0},
			deltaTime:   1.0,
			expectedPos: types.Vector2{X: 100, Y: 0},
		},
		{
			name:        "Move Diagonal",
			startPos:    types.Vector2{X: 0, Y: 0},
			velocity:    types.Vector2{X: 100, Y: 50},
			deltaTime:   2.0,
			expectedPos: types.Vector2{X: 200, Y: 100},
		},
		{
			name:        "Negative Velocity",
			startPos:    types.Vector2{X: 100, Y: 100},
			velocity:    types.Vector2{X: -50, Y: -25},
			deltaTime:   1.0,
			expectedPos: types.Vector2{X: 50, Y: 75},
		},
		{
			name:        "Small DeltaTime",
			startPos:    types.Vector2{X: 0, Y: 0},
			velocity:    types.Vector2{X: 100, Y: 100},
			deltaTime:   0.016, // ~60fps
			expectedPos: types.Vector2{X: 1.6, Y: 1.6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mover := NewBasicMover(tt.startPos, tt.velocity, 64, 64)
			mover.Update(tt.deltaTime)

			got := mover.GetPosition()
			if !floatEquals(got.X, tt.expectedPos.X, 0.0001) || !floatEquals(got.Y, tt.expectedPos.Y, 0.0001) {
				t.Errorf("Expected position %v, got %v", tt.expectedPos, got)
			}
		})
	}
}

func TestBasicMoverScreenWrappingRight(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 750, Y: 300},
		types.Vector2{X: 100, Y: 0},
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	// Move past right edge
	mover.Update(1.0) // Position becomes 850

	got := mover.GetPosition()
	// Should wrap to -64 (just off left side)
	if got.X != -64 {
		t.Errorf("Expected X position -64 after wrapping right, got %v", got.X)
	}
}

func TestBasicMoverScreenWrappingLeft(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 10, Y: 300},
		types.Vector2{X: -100, Y: 0},
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	// Move past left edge
	mover.Update(0.5) // Position becomes -40, but sprite extends 64 pixels
	// Not wrapped yet (sprite still partially visible)

	mover.Update(0.5) // Position becomes -90, sprite fully off screen
	got := mover.GetPosition()

	// Should wrap to right side (800)
	if got.X != 800 {
		t.Errorf("Expected X position 800 after wrapping left, got %v", got.X)
	}
}

func TestBasicMoverScreenWrappingDown(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 300, Y: 550},
		types.Vector2{X: 0, Y: 100},
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	// Move past bottom edge
	mover.Update(1.0) // Position becomes 650

	got := mover.GetPosition()
	// Should wrap to -64 (just off top)
	if got.Y != -64 {
		t.Errorf("Expected Y position -64 after wrapping down, got %v", got.Y)
	}
}

func TestBasicMoverScreenWrappingUp(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 300, Y: 10},
		types.Vector2{X: 0, Y: -100},
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	// Move past top edge
	mover.Update(0.5) // Position becomes -40
	mover.Update(0.5) // Position becomes -90, sprite fully off screen

	got := mover.GetPosition()
	// Should wrap to bottom (600)
	if got.Y != 600 {
		t.Errorf("Expected Y position 600 after wrapping up, got %v", got.Y)
	}
}

func TestBasicMoverNoWrappingWhenStationary(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 900, Y: 700}, // Outside bounds
		types.Vector2{X: 0, Y: 0},     // But not moving
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	mover.Update(1.0)

	got := mover.GetPosition()
	// Should not wrap if not moving
	if got.X != 900 || got.Y != 700 {
		t.Errorf("Stationary object wrapped unexpectedly: %v", got)
	}
}

func TestBasicMoverSetScreenBounds(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 0, Y: 0},
		types.Vector2{X: 0, Y: 0},
		64, 64,
	)

	// Should accept any positive values without error
	mover.SetScreenBounds(1920, 1080)
	mover.SetScreenBounds(640, 480)
	mover.SetScreenBounds(1000, 1000)

	// Test doesn't crash - success
}

func TestBasicMoverMultipleUpdates(t *testing.T) {
	mover := NewBasicMover(
		types.Vector2{X: 0, Y: 0},
		types.Vector2{X: 10, Y: 10},
		64, 64,
	)

	// Simulate 10 frames at 60fps
	for i := 0; i < 10; i++ {
		mover.Update(0.016)
	}

	got := mover.GetPosition()
	expected := types.Vector2{X: 1.6, Y: 1.6}

	if !floatEquals(got.X, expected.X, 0.001) || !floatEquals(got.Y, expected.Y, 0.001) {
		t.Errorf("After 10 updates, expected position %v, got %v", expected, got)
	}
}

// Helper function to compare floats with tolerance
func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}

func BenchmarkBasicMoverUpdate(b *testing.B) {
	mover := NewBasicMover(
		types.Vector2{X: 100, Y: 100},
		types.Vector2{X: 50, Y: 50},
		64, 64,
	)
	mover.SetScreenBounds(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mover.Update(0.016)
	}
}

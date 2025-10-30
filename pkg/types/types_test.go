package types

import (
	"testing"
)

func TestVector2(t *testing.T) {
	v := Vector2{X: 10.5, Y: 20.75}

	if v.X != 10.5 {
		t.Errorf("Expected X=10.5, got %f", v.X)
	}
	if v.Y != 20.75 {
		t.Errorf("Expected Y=20.75, got %f", v.Y)
	}
}

func TestVector2ZeroValue(t *testing.T) {
	var v Vector2

	if v.X != 0 || v.Y != 0 {
		t.Errorf("Expected zero vector, got (%f, %f)", v.X, v.Y)
	}
}

func TestUVRect(t *testing.T) {
	uv := UVRect{
		U: 0.0,
		V: 0.5,
		W: 0.25,
		H: 0.25,
	}

	if uv.U != 0.0 || uv.V != 0.5 || uv.W != 0.25 || uv.H != 0.25 {
		t.Errorf("UV rect values incorrect: %+v", uv)
	}
}

func TestSpriteRenderData(t *testing.T) {
	renderData := SpriteRenderData{
		TexturePath: "test.png",
		Position:    Vector2{X: 100, Y: 200},
		Size:        Vector2{X: 64, Y: 64},
		UV: UVRect{
			U: 0.0,
			V: 0.0,
			W: 1.0,
			H: 1.0,
		},
		Visible: true,
		Frame:   0,
	}

	if renderData.TexturePath != "test.png" {
		t.Errorf("Expected texture path 'test.png', got '%s'", renderData.TexturePath)
	}
	if renderData.Position.X != 100 {
		t.Errorf("Expected X=100, got %f", renderData.Position.X)
	}
	if !renderData.Visible {
		t.Error("Expected visible to be true")
	}
}

func TestObjectState(t *testing.T) {
	state := ObjectState{
		ID:       "test-id",
		Position: Vector2{X: 50, Y: 75},
		Visible:  true,
		Frame:    3,
	}

	if state.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", state.ID)
	}
	if state.Position.X != 50 || state.Position.Y != 75 {
		t.Errorf("Expected position (50, 75), got (%f, %f)", state.Position.X, state.Position.Y)
	}
	if !state.Visible {
		t.Error("Expected visible to be true")
	}
	if state.Frame != 3 {
		t.Errorf("Expected frame 3, got %d", state.Frame)
	}
}

func TestCopyObjectState(t *testing.T) {
	original := ObjectState{
		ID:       "original",
		Position: Vector2{X: 100, Y: 200},
		Visible:  true,
		Frame:    5,
	}

	copy := CopyObjectState(original)

	// Verify copy has same values
	if copy.ID != original.ID {
		t.Errorf("Expected ID '%s', got '%s'", original.ID, copy.ID)
	}
	if copy.Position != original.Position {
		t.Errorf("Expected position %v, got %v", original.Position, copy.Position)
	}

	// Modify original - copy should be unchanged
	original.ID = "modified"
	original.Position.X = 999

	if copy.ID != "original" {
		t.Error("Copy was modified when original changed")
	}
	if copy.Position.X != 100 {
		t.Error("Copy position was modified when original changed")
	}
}

func TestInputState(t *testing.T) {
	state := InputState{
		MoveUp:    true,
		MoveDown:  false,
		MoveLeft:  true,
		MoveRight: false,
	}

	if !state.MoveUp {
		t.Error("Expected MoveUp to be true")
	}
	if state.MoveDown {
		t.Error("Expected MoveDown to be false")
	}
	if !state.MoveLeft {
		t.Error("Expected MoveLeft to be true")
	}
	if state.MoveRight {
		t.Error("Expected MoveRight to be false")
	}
}

func TestInputStateZeroValue(t *testing.T) {
	var state InputState

	if state.MoveUp || state.MoveDown || state.MoveLeft || state.MoveRight {
		t.Error("Expected all zero values to be false")
	}
}

func TestGameStateString(t *testing.T) {
	tests := []struct {
		state    GameState
		expected string
	}{
		{GAMEPLAY, "SPRITE"},
		{TRIANGLE, "TRIANGLE"},
	}

	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, got)
		}
	}
}

func TestPipelineTypeString(t *testing.T) {
	tests := []struct {
		pipeline PipelineType
		expected string
	}{
		{TrianglePipeline, "TrianglePipeline"},
		{TexturedPipeline, "TexturedPipeline"},
	}

	for _, tt := range tests {
		got := tt.pipeline.String()
		if got != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, got)
		}
	}
}

func TestSpriteInterface(t *testing.T) {
	// This test verifies that the Sprite interface is correctly defined
	// We can't instantiate an interface, but we can check it compiles
	var _ Sprite = nil // Should compile
}

func TestMoverInterface(t *testing.T) {
	// This test verifies that the Mover interface is correctly defined
	var _ Mover = nil // Should compile
}

func TestGameObjectInterface(t *testing.T) {
	// This test verifies that the GameObject interface is correctly defined
	var _ GameObject = nil // Should compile
}

func TestInputCapturerInterface(t *testing.T) {
	// This test verifies that the InputCapturer interface is correctly defined
	var _ InputCapturer = nil // Should compile
}

func BenchmarkCopyObjectState(b *testing.B) {
	state := ObjectState{
		ID:       "test",
		Position: Vector2{X: 100, Y: 200},
		Visible:  true,
		Frame:    5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyObjectState(state)
	}
}

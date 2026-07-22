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
		{MENU, "MENU"},
		{GAMEPLAY, "GAMEPLAY"},
		{PLAYER_MENU, "PLAYER_MENU"},
		{BATTLE, "BATTLE"},
	}

	for _, tt := range tests {
		got := tt.state.String()
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

func TestInputCapturerInterface(t *testing.T) {
	// This test verifies that the InputCapturer interface is correctly defined
	var _ InputCapturer = nil // Should compile
}

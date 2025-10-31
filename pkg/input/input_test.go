package input

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func TestNewMockInput(t *testing.T) {
	mockInput := NewMockInput()

	if mockInput == nil {
		t.Fatal("NewMockInput returned nil")
	}

	state := mockInput.GetInputState()
	if state.MoveUp || state.MoveDown || state.MoveLeft || state.MoveRight {
		t.Error("Expected all input states to be false initially")
	}
}

func TestMockInputSetGet(t *testing.T) {
	mockInput := NewMockInput()

	newState := types.InputState{
		MoveUp:    true,
		MoveDown:  false,
		MoveLeft:  true,
		MoveRight: false,
	}

	mockInput.SetInputState(newState)

	got := mockInput.GetInputState()
	if got.MoveUp != newState.MoveUp ||
		got.MoveDown != newState.MoveDown ||
		got.MoveLeft != newState.MoveLeft ||
		got.MoveRight != newState.MoveRight {
		t.Errorf("Expected state %+v, got %+v", newState, got)
	}
}

func TestMockInputInitialize(t *testing.T) {
	mockInput := NewMockInput()

	if mockInput.IsInitialized() {
		t.Error("Expected mock to not be initialized initially")
	}

	err := mockInput.Initialize()
	if err != nil {
		t.Errorf("Initialize returned error: %v", err)
	}

	if !mockInput.IsInitialized() {
		t.Error("Expected mock to be initialized after Initialize()")
	}
}

func TestMockInputCleanup(t *testing.T) {
	mockInput := NewMockInput()
	mockInput.Initialize()

	mockInput.SetInputState(types.InputState{MoveUp: true, MoveRight: true})

	mockInput.Cleanup()

	if mockInput.IsInitialized() {
		t.Error("Expected mock to not be initialized after Cleanup()")
	}

	state := mockInput.GetInputState()
	if state.MoveUp || state.MoveRight {
		t.Error("Expected state to be reset after Cleanup()")
	}
}

func TestMockInputConcurrency(t *testing.T) {
	mockInput := NewMockInput()
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mockInput.SetInputState(types.InputState{MoveUp: i%2 == 0})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = mockInput.GetInputState()
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done
}

func TestInputStateAllDirections(t *testing.T) {
	tests := []struct {
		name  string
		state types.InputState
	}{
		{
			name:  "No Movement",
			state: types.InputState{},
		},
		{
			name:  "Up Only",
			state: types.InputState{MoveUp: true},
		},
		{
			name:  "Down Only",
			state: types.InputState{MoveDown: true},
		},
		{
			name:  "Left Only",
			state: types.InputState{MoveLeft: true},
		},
		{
			name:  "Right Only",
			state: types.InputState{MoveRight: true},
		},
		{
			name: "Up-Right",
			state: types.InputState{
				MoveUp:    true,
				MoveRight: true,
			},
		},
		{
			name: "All Directions",
			state: types.InputState{
				MoveUp:    true,
				MoveDown:  true,
				MoveLeft:  true,
				MoveRight: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInput := NewMockInput()
			mockInput.SetInputState(tt.state)

			got := mockInput.GetInputState()
			if got != tt.state {
				t.Errorf("Expected state %+v, got %+v", tt.state, got)
			}
		})
	}
}

func TestMockInputMultipleInitialize(t *testing.T) {
	mockInput := NewMockInput()

	// Multiple initializations should not error
	mockInput.Initialize()
	mockInput.Initialize()
	mockInput.Initialize()

	if !mockInput.IsInitialized() {
		t.Error("Expected mock to remain initialized")
	}
}

func TestMockInputMultipleCleanup(t *testing.T) {
	mockInput := NewMockInput()
	mockInput.Initialize()

	// Multiple cleanups should not error
	mockInput.Cleanup()
	mockInput.Cleanup()
	mockInput.Cleanup()

	if mockInput.IsInitialized() {
		t.Error("Expected mock to remain not initialized")
	}
}

func TestMockInputStateTransitions(t *testing.T) {
	mockInput := NewMockInput()

	// Test various state transitions
	states := []types.InputState{
		{MoveUp: true},
		{MoveRight: true},
		{MoveDown: true},
		{MoveLeft: true},
		{MoveUp: true, MoveRight: true},
		{},
		{MoveDown: true, MoveLeft: true},
		{MoveUp: true, MoveDown: true, MoveLeft: true, MoveRight: true},
		{},
	}

	for i, state := range states {
		mockInput.SetInputState(state)
		got := mockInput.GetInputState()

		if got != state {
			t.Errorf("Transition %d: expected %+v, got %+v", i, state, got)
		}
	}
}

func BenchmarkMockInputGetState(b *testing.B) {
	mockInput := NewMockInput()
	mockInput.SetInputState(types.InputState{MoveUp: true, MoveRight: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockInput.GetInputState()
	}
}

func BenchmarkMockInputSetState(b *testing.B) {
	mockInput := NewMockInput()
	state := types.InputState{MoveUp: true, MoveRight: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockInput.SetInputState(state)
	}
}

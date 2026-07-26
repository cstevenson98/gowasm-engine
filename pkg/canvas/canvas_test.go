package canvas

import (
	"testing"

	"github.com/cstevenson98/milo/pkg/types"
)

func TestMockCanvasManager_Initialize(t *testing.T) {
	mock := NewMockCanvasManager()
	if mock.IsInitialized() {
		t.Fatalf("expected mock to start uninitialized")
	}

	if err := mock.Initialize("test-canvas"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.IsInitialized() {
		t.Errorf("expected mock to be initialized after Initialize")
	}
}

func TestMockCanvasManager_LoadAndDraw(t *testing.T) {
	mock := NewMockCanvasManager()

	// Drawing before initialization should fail.
	if err := mock.DrawTexturedRect("tex.png", types.Vector2{}, types.Vector2{}, types.UVRect{}); err == nil {
		t.Errorf("expected error drawing before initialization")
	}

	mock.Initialize("test-canvas")

	// Drawing an unloaded texture should fail.
	if err := mock.DrawTexturedRect("tex.png", types.Vector2{}, types.Vector2{}, types.UVRect{}); err == nil {
		t.Errorf("expected error drawing unloaded texture")
	}

	if err := mock.LoadTexture("tex.png"); err != nil {
		t.Fatalf("unexpected error loading texture: %v", err)
	}

	if err := mock.DrawTexturedRect("tex.png", types.Vector2{X: 1}, types.Vector2{X: 2}, types.UVRect{}); err != nil {
		t.Fatalf("unexpected error drawing loaded texture: %v", err)
	}

	if mock.DrawnRectCount() != 1 {
		t.Errorf("expected 1 draw call, got %d", mock.DrawnRectCount())
	}
}

func TestMockCanvasManager_Cleanup(t *testing.T) {
	mock := NewMockCanvasManager()
	mock.Initialize("test-canvas")

	if err := mock.Cleanup(); err != nil {
		t.Fatalf("unexpected error during cleanup: %v", err)
	}
	if !mock.WasCleanupCalled() {
		t.Errorf("expected cleanup to be recorded")
	}
	if mock.IsInitialized() {
		t.Errorf("expected canvas uninitialized after cleanup")
	}
}

func TestCanvasError(t *testing.T) {
	err := &CanvasError{Message: "Test error"}
	if err.Error() != "Test error" {
		t.Errorf("expected 'Test error', got: %s", err.Error())
	}
}

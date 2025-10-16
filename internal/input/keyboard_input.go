//go:build js

package input

import (
	"sync"
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/types"
)

// KeyboardInput captures keyboard input from the browser
type KeyboardInput struct {
	inputState  types.InputState
	mu          sync.RWMutex
	keydownFunc js.Func
	keyupFunc   js.Func
	initialized bool
}

// NewKeyboardInput creates a new keyboard input capturer
func NewKeyboardInput() *KeyboardInput {
	return &KeyboardInput{
		inputState:  types.InputState{},
		initialized: false,
	}
}

// GetInputState returns the current input state
func (k *KeyboardInput) GetInputState() types.InputState {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.inputState
}

// Initialize sets up keyboard event listeners
func (k *KeyboardInput) Initialize() error {
	if k.initialized {
		return nil
	}

	// Create keydown handler
	k.keydownFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		event := args[0]
		key := event.Get("key").String()

		k.mu.Lock()
		defer k.mu.Unlock()

		switch key {
		case "w", "W":
			k.inputState.MoveUp = true
		case "s", "S":
			k.inputState.MoveDown = true
		case "a", "A":
			k.inputState.MoveLeft = true
		case "d", "D":
			k.inputState.MoveRight = true
		}

		return nil
	})

	// Create keyup handler
	k.keyupFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		event := args[0]
		key := event.Get("key").String()

		k.mu.Lock()
		defer k.mu.Unlock()

		switch key {
		case "w", "W":
			k.inputState.MoveUp = false
		case "s", "S":
			k.inputState.MoveDown = false
		case "a", "A":
			k.inputState.MoveLeft = false
		case "d", "D":
			k.inputState.MoveRight = false
		}

		return nil
	})

	// Add event listeners to the document
	js.Global().Get("document").Call("addEventListener", "keydown", k.keydownFunc)
	js.Global().Get("document").Call("addEventListener", "keyup", k.keyupFunc)

	k.initialized = true
	println("DEBUG: Keyboard input initialized - Use WASD to move")

	return nil
}

// Cleanup releases keyboard event listeners
func (k *KeyboardInput) Cleanup() {
	if !k.initialized {
		return
	}

	js.Global().Get("document").Call("removeEventListener", "keydown", k.keydownFunc)
	js.Global().Get("document").Call("removeEventListener", "keyup", k.keyupFunc)

	k.keydownFunc.Release()
	k.keyupFunc.Release()

	k.initialized = false
	println("DEBUG: Keyboard input cleaned up")
}

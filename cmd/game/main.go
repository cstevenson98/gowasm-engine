//go:build js

package main

import (
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/engine"
	"github.com/conor/webgpu-triangle/internal/logger"
)

// Global engine instance
var gameEngine *engine.Engine

func main() {
	logger.Logger.Info("Go WASM program started")

	// Create the game engine
	gameEngine = engine.NewEngine()

	// Check if DOM is already loaded
	document := js.Global().Get("document")
	if document.Get("readyState").String() == "complete" {
		logger.Logger.Info("DOM already loaded, initializing immediately")
		initializeEngine()
	} else {
		logger.Logger.Info("Waiting for DOM to load")
		// Wait for DOM to be ready
		js.Global().Call("addEventListener", "DOMContentLoaded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			logger.Logger.Info("DOMContentLoaded event fired")
			initializeEngine()
			return nil
		}))
	}

	// Keep the program running
	<-make(chan bool)
}

func initializeEngine() {
	logger.Logger.Info("Starting engine initialization")

	err := gameEngine.Initialize("webgpu-canvas")
	if err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return
	}

	// Start the game loop
	gameEngine.Start()
}

//go:build js

package main

import (
	"syscall/js"

	exts "example.com/basic-game/scenes"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/engine"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
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

// createCanvas creates the game canvas element and adds it to the DOM
func createCanvas() string {
	document := js.Global().Get("document")

	// Create canvas element
	canvas := document.Call("createElement", "canvas")
	canvas.Set("id", "webgpu-canvas")
	canvas.Set("width", config.Global.Screen.CanvasWidth)
	canvas.Set("height", config.Global.Screen.CanvasHeight)

	// Add canvas to game-container
	container := document.Call("getElementById", "game-container")
	if container.IsNull() || container.IsUndefined() {
		logger.Logger.Error("game-container element not found")
		return ""
	}

	container.Call("appendChild", canvas)

	logger.Logger.Infof("Canvas created: %dx%d", config.Global.Screen.CanvasWidth, config.Global.Screen.CanvasHeight)
	return "webgpu-canvas"
}

func initializeEngine() {
	logger.Logger.Info("Starting engine initialization")

	// Create the canvas element
	canvasID := createCanvas()
	if canvasID == "" {
		logger.Logger.Error("Failed to create canvas")
		return
	}

	// Create and register scenes (demonstrating library usage)
	// Input capturer will be injected by the engine during scene initialization

	// Create gameplay scene
	gameplayScene := exts.NewGameplayScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	gameplayScene.SetCanvasManager(gameEngine.GetCanvasManager())
	err := gameplayScene.InitializeDebugConsole()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize debug console for gameplay: %s", err)
	}

	// Create battle scene
	battleScene := exts.NewBattleScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	battleScene.SetCanvasManager(gameEngine.GetCanvasManager())
	err = battleScene.InitializeDebugConsole()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize debug console for battle: %s", err)
	}

	// Register both scenes with the engine
	gameEngine.RegisterScene(types.GAMEPLAY, gameplayScene)
	gameEngine.RegisterScene(types.BATTLE, battleScene)

	logger.Logger.Info("Scenes registered: Press 1 to switch to gameplay, 2 to switch to battle")

	// Initialize the engine
	err = gameEngine.Initialize(canvasID)
	if err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return
	}

	// Set the initial game state
	err = gameEngine.SetGameState(types.GAMEPLAY)
	if err != nil {
		logger.Logger.Errorf("Failed to set initial game state: %s", err.Error())
		return
	}

	// Start the game loop
	gameEngine.Start()
}

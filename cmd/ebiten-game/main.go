package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"example.com/basic-game/game/gamestate"
	exts "example.com/basic-game/scenes"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/engine"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func main() {
	logger.Logger.Info("Ebiten game starting")

	// Create the Ebiten game engine
	gameEngine := engine.NewEbitenEngine()

	// Create game state manager
	stateManager := gamestate.NewGameStateManager()

	// Register game state manager with engine
	gameEngine.RegisterGameStateProvider(stateManager)

	// Create and register scenes
	menuScene := exts.NewMenuScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	menuScene.SetCanvasManager(gameEngine.GetCanvasManager())

	gameplayScene := exts.NewGameplayScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	gameplayScene.SetCanvasManager(gameEngine.GetCanvasManager())

	battleScene := exts.NewBattleScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	battleScene.SetCanvasManager(gameEngine.GetCanvasManager())

	playerMenuScene := exts.NewPlayerMenuScene(
		config.Global.Screen.Width,
		config.Global.Screen.Height,
	)
	playerMenuScene.SetCanvasManager(gameEngine.GetCanvasManager())

	// Register all scenes with the engine
	gameEngine.RegisterScene(types.MENU, menuScene)
	gameEngine.RegisterScene(types.GAMEPLAY, gameplayScene)
	gameEngine.RegisterScene(types.PLAYER_MENU, playerMenuScene)
	gameEngine.RegisterScene(types.BATTLE, battleScene)

	logger.Logger.Info("Scenes registered: Menu, Gameplay, PlayerMenu, Battle")

	// Initialize the engine
	err := gameEngine.Initialize("ebiten-canvas") // Canvas ID not used by Ebiten
	if err != nil {
		log.Fatalf("Engine initialization failed: %s", err.Error())
	}

	// Set the initial game state to menu
	err = gameEngine.SetGameState(types.MENU)
	if err != nil {
		log.Fatalf("Failed to set initial game state: %s", err.Error())
	}

	// Start the engine (sets running flag)
	gameEngine.Start()

	// Configure Ebiten window
	windowWidth := int(config.Global.Screen.Width * float64(config.Global.Rendering.PixelScale))
	windowHeight := int(config.Global.Screen.Height * float64(config.Global.Rendering.PixelScale))
	
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Go WASM Engine - Ebiten Edition")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	// Disable screen filter for pixel-perfect rendering
	if config.Global.Rendering.PixelArtMode {
		ebiten.SetScreenFilterEnabled(false)
	}

	logger.Logger.Infof("Starting Ebiten game loop (window: %dx%d, virtual: %.0fx%.0f)", 
		windowWidth, windowHeight, config.Global.Screen.Width, config.Global.Screen.Height)

	// Run the game - this blocks until the game ends
	if err := ebiten.RunGame(gameEngine); err != nil {
		log.Fatal(err)
	}

	logger.Logger.Info("Game ended")
}

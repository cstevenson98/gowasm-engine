package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"example.com/basic-game/game/gamestate"
	"example.com/basic-game/states"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/engine"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func main() {
	logger.Logger.Info("Ebiten game starting")

	// Create the Ebiten game engine
	gameEngine := engine.NewEngine()

	// Create game state manager
	stateManager := gamestate.NewGameStateManager()

	// Register game state manager with engine
	gameEngine.RegisterGameStateProvider(stateManager)

	// Create and register states. The engine injects dependencies (input, UI,
	// screen size, state-change callback, game-state provider) into each state
	// when it becomes active, so no manual wiring is needed here.
	gameEngine.RegisterState(types.MENU, states.NewMenuState())
	gameEngine.RegisterState(types.GAMEPLAY, states.NewGameplayState())
	gameEngine.RegisterState(types.PLAYER_MENU, states.NewPlayerMenuState())
	gameEngine.RegisterState(types.BATTLE, states.NewBattleState())

	logger.Logger.Info("States registered: Menu, Gameplay, PlayerMenu, Battle")

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

	// Configure Ebiten window. The window size is derived from the virtual
	// resolution and the pixel scale.
	windowWidth := config.WindowWidth()
	windowHeight := config.WindowHeight()

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

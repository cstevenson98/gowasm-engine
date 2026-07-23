// Command game is the browser (WASM) entry point for the basic-game example.
// It mirrors cmd/ebiten-game (the desktop entry): both register the same states
// and hand control to Ebiten, which drives either the desktop window or the
// browser canvas depending on the build target (GOOS=js GOARCH=wasm).
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
	logger.Logger.Info("Ebiten game starting (wasm entry)")

	cfg := config.Default()
	gameEngine := engine.NewEngine(cfg)

	stateManager := gamestate.NewGameStateManager()
	gameEngine.RegisterGameStateProvider(stateManager)

	// Register states. The engine injects dependencies (input, UI, screen size,
	// state-change callback, game-state provider) into each state on activation.
	gameEngine.RegisterState(types.MENU, states.NewMenuState())
	gameEngine.RegisterState(types.GAMEPLAY, states.NewGameplayState())
	gameEngine.RegisterState(types.PLAYER_MENU, states.NewPlayerMenuState())
	gameEngine.RegisterState(types.BATTLE, states.NewBattleState())

	logger.Logger.Info("States registered: Menu, Gameplay, PlayerMenu, Battle")

	if err := gameEngine.Initialize("ebiten-canvas"); err != nil {
		log.Fatalf("Engine initialization failed: %s", err.Error())
	}

	if err := gameEngine.SetGameState(types.MENU); err != nil {
		log.Fatalf("Failed to set initial game state: %s", err.Error())
	}

	gameEngine.Start()

	ebiten.SetWindowSize(cfg.WindowWidth(), cfg.WindowHeight())
	ebiten.SetWindowTitle("Go WASM Engine - Ebiten Edition")
	if cfg.Rendering.PixelArtMode {
		ebiten.SetScreenFilterEnabled(false)
	}

	if err := ebiten.RunGame(gameEngine); err != nil {
		log.Fatal(err)
	}

	logger.Logger.Info("Game ended")
}

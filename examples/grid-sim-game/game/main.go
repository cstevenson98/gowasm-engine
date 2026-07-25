// Command game is the entry point for the grid-sim-game example. It builds
// for both the browser (GOOS=js GOARCH=wasm, via examples/Makefile) and
// desktop (`go run ./game`), same as examples/basic-game/game.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"example.com/grid-sim-game/states"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/engine"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func main() {
	logger.Logger.Info("Grid sim game starting")

	// 1080p 16:9 virtual resolution. The world grid is much larger than the
	// viewport (see gameconfig.GridCols/Rows); explore with WASD / middle-mouse
	// pan. The right half of the window is the ImGui network panel.
	cfg := config.Default()
	cfg.Screen.Width = 1920
	cfg.Screen.Height = 1080
	cfg.Rendering.PixelScale = 1
	// Opt into ImGui for the right-half network inspector panel.
	gameEngine := engine.NewEngine(cfg).EnableImGui()

	// This example has a single state (the grid itself), registered under the
	// engine's existing GAMEPLAY value - no menu/battle states needed.
	gameEngine.RegisterState(types.GAMEPLAY, states.NewGridState())

	if err := gameEngine.Initialize("grid-canvas"); err != nil {
		log.Fatalf("Engine initialization failed: %s", err.Error())
	}

	if err := gameEngine.SetGameState(types.GAMEPLAY); err != nil {
		log.Fatalf("Failed to set initial game state: %s", err.Error())
	}

	gameEngine.Start()

	ebiten.SetWindowSize(cfg.WindowWidth(), cfg.WindowHeight())
	ebiten.SetWindowTitle("Grid Sim Game")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if cfg.Rendering.PixelArtMode {
		ebiten.SetScreenFilterEnabled(false)
	}

	if err := ebiten.RunGame(gameEngine); err != nil {
		log.Fatal(err)
	}

	logger.Logger.Info("Grid sim game ended")
}

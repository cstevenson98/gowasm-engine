// Command game is a minimal illustrative demo for milo: a counter
// shown as large centered text on a 1280×720 screen; Up arrow increments it.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"example.com/demo/states"
	"github.com/cstevenson98/milo/pkg/config"
	"github.com/cstevenson98/milo/pkg/engine"
	"github.com/cstevenson98/milo/pkg/logger"
	"github.com/cstevenson98/milo/pkg/types"
)

func main() {
	logger.Logger.Info("milo demo starting")

	cfg := config.Default()
	cfg.Screen.Width = 1280
	cfg.Screen.Height = 720
	cfg.Rendering.PixelScale = 1

	eng := engine.NewEngine(cfg)
	eng.RegisterState(types.GAMEPLAY, states.NewCounterState())

	if err := eng.Initialize("ebiten-canvas"); err != nil {
		log.Fatalf("engine init: %v", err)
	}
	if err := eng.SetGameState(types.GAMEPLAY); err != nil {
		log.Fatalf("set state: %v", err)
	}

	eng.Start()

	ebiten.SetWindowSize(cfg.WindowWidth(), cfg.WindowHeight())
	ebiten.SetWindowTitle("milo — counter demo")
	if cfg.Rendering.PixelArtMode {
		ebiten.SetScreenFilterEnabled(false)
	}

	if err := ebiten.RunGame(eng); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game implements ebiten.Game interface
type Game struct {
	llama *ebiten.Image
	x, y  float64
}

// Update updates the game logic (called 60 times per second)
func (g *Game) Update() error {
	// Quit on ESC
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Move llama with arrow keys
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.x += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.x -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.y += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.y -= 2
	}

	return nil
}

// Draw draws the game screen (called every frame)
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen to black
	screen.Clear()

	// Draw the llama
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(g.x, g.y)
	
	// Use nearest-neighbor filtering for pixel-perfect rendering
	opts.Filter = ebiten.FilterNearest
	
	screen.DrawImage(g.llama, opts)

	// Show instructions
	ebitenutil.DebugPrint(screen, "Arrow keys to move llama\nESC to quit")
}

// Layout returns the game's logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}

func main() {
	// Load the llama image
	llama, _, err := ebitenutil.NewImageFromFile("../basic-game/assets/llama.png")
	if err != nil {
		log.Fatalf("Failed to load llama.png: %v", err)
	}

	// Create game
	game := &Game{
		llama: llama,
		x:     384, // Center horizontally (800-32)/2
		y:     284, // Center vertically (600-32)/2
	}

	// Set window properties (3x scaling: 800*3 = 2400, 600*3 = 1800)
	ebiten.SetWindowSize(2400, 1800)
	ebiten.SetWindowTitle("Ebiten Llama Demo (3x Pixel Perfect)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	// Disable screen filter for pixel-perfect rendering
	ebiten.SetScreenFilterEnabled(false)

	// Run the game
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

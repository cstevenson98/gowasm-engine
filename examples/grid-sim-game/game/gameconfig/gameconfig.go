// Package gameconfig holds configuration specific to the grid-sim-game
// example: grid dimensions, tile/camera sizing, toolbar layout, and texture
// paths. Mirrors examples/basic-game/game/gameconfig in spirit: this configures
// one specific game, not the reusable engine, so a package-level global is the
// simplest fit.
package gameconfig

// Settings contains this game's configuration.
type Settings struct {
	TileSize    float64 // Width/height of one grid cell, in virtual pixels.
	GridCols    int     // Number of columns in the grid.
	GridRows    int     // Number of rows in the grid.
	CameraSpeed float64 // Camera scroll speed, in pixels per second.

	ToolbarHeight float64 // Height of the top toolbar strip, in virtual pixels.
	ButtonWidth   float64 // Width of each toolbar button.
	ButtonHeight  float64 // Height of each toolbar button.
	ButtonGap     float64 // Horizontal gap between toolbar buttons.
	ButtonMarginX float64 // Left margin before the first toolbar button.
	ButtonMarginY float64 // Top margin of each toolbar button within the bar.

	BlankTexture     string // Path to the single-cell blank/grid tile texture.
	GeneratorTexture string // Path to the generator tile texture.
	HouseTexture     string // Path to the house tile texture.
	LineTexture      string // Path to the line tile texture.
}

// Global is this game's configuration singleton.
var Global = Settings{
	TileSize:    32.0,
	GridCols:    20,
	GridRows:    20,
	CameraSpeed: 150.0,

	ToolbarHeight: 20.0,
	ButtonWidth:   70.0,
	ButtonHeight:  16.0,
	ButtonGap:     6.0,
	ButtonMarginX: 6.0,
	ButtonMarginY: 2.0,

	BlankTexture:     "assets/art/blank.png",
	GeneratorTexture: "assets/art/generator.png",
	HouseTexture:     "assets/art/house.png",
	LineTexture:      "assets/art/line.png",
}

// WorldWidth returns the total grid width in virtual pixels.
func (s Settings) WorldWidth() float64 { return float64(s.GridCols) * s.TileSize }

// WorldHeight returns the total grid height in virtual pixels.
func (s Settings) WorldHeight() float64 { return float64(s.GridRows) * s.TileSize }

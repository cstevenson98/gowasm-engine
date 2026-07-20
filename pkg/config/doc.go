// Package config centralises all tunable settings for the engine and game in
// one place.
//
// Rather than scattering magic numbers (screen size, player speed, animation
// timings, debug colours, battle balance) across the codebase, they are
// gathered into the Settings struct and exposed through the package-level
// Global variable. Any package can read Global to stay in sync, and a game can
// override fields on Global at startup to retune behaviour without touching
// engine code.
//
// Settings is grouped by concern - Screen, Player, Animation, Debug, Rendering,
// and Battle - each documented field by field. Global ships with sensible
// defaults for the bundled example game.
//
// # Resolution and scaling
//
// The engine draws to a fixed virtual resolution (Screen.Width x Screen.Height)
// which Ebiten scales to the actual window. Rendering.PixelScale and the
// pixel-art flags control crisp integer upscaling for pixel-art visuals.
package config

// Package config defines the engine's own configuration: virtual screen
// resolution, rendering quality, and debug-console behaviour/appearance.
//
// There is deliberately no package-level mutable global here. Construct a
// Settings value (or start from Default()) and pass it to engine.NewEngine;
// the engine threads the relevant pieces down into canvas/ui/text/state, so
// every consumer sees the same, explicitly-chosen configuration rather than
// reaching into a shared singleton.
//
// Game-specific configuration (player stats, enemy content, save-file
// defaults, ...) does not belong here - it lives in the game itself, e.g.
// the sibling rpg-game module's game/gameconfig.
package config

// Settings is the engine's configuration.
type Settings struct {
	Screen    ScreenSettings
	Debug     DebugSettings
	Rendering RenderingSettings
	Animation AnimationSettings
}

// ScreenSettings contains display configuration. The virtual resolution is the
// fixed coordinate space the game renders in; the actual window size is derived
// from it and Rendering.PixelScale (see Settings.WindowWidth/WindowHeight).
type ScreenSettings struct {
	Width  float64 // Virtual game resolution width
	Height float64 // Virtual game resolution height
}

// AnimationSettings contains engine-level animation timing defaults.
type AnimationSettings struct {
	// DefaultFrameTime is the fallback seconds-per-frame used by generic
	// prefab helpers (e.g. prefab.NewLlama) when a game doesn't specify one.
	DefaultFrameTime float64
}

// DebugSettings contains debug console configuration
type DebugSettings struct {
	Enabled                   bool       // Enable/disable debug console
	FontPath                  string     // Path to font sprite sheet (without .sheet.png extension)
	FontScale                 float64    // Scale factor for debug text (1.0 = normal, 2.0 = double size)
	CharacterSpacingReduction float64    // Pixels to reduce character spacing (reduces padding between letters)
	MaxMessages               int        // Maximum number of messages to display
	MessageLifetime           float64    // Time before messages fade out (0 = never fade)
	ConsoleHeight             float64    // Height of the console in pixels
	BackgroundColor           [4]float32 // RGBA background color (with alpha for transparency)
	TextColor                 [4]float32 // RGBA text color
}

// RenderingSettings contains rendering quality and style configuration
type RenderingSettings struct {
	PixelArtMode        bool    // Enable pixel-perfect rendering (nearest-neighbor filtering)
	TextureFiltering    string  // "nearest" or "linear" - texture filtering mode
	PixelPerfectScaling bool    // Ensure integer scaling for pixel art
	PixelScale          int     // Real pixels per game pixel (e.g., 4 = 4x4 pixels per game pixel)
	UILineSpacing       float64 // Line spacing multiplier for UI elements (menus, logs, status)
	TextLineSpacing     float64 // Line spacing multiplier for paragraph text (newlines within strings)
}

// Default returns the engine's stock settings. Games that don't need custom
// values can pass this straight to engine.NewEngine; games that do can start
// from it and override individual fields.
func Default() Settings {
	return Settings{
		Screen: ScreenSettings{
			Width:  320.0, // Virtual game resolution (240p, 4:3)
			Height: 240.0,
		},
		Animation: AnimationSettings{
			DefaultFrameTime: 0.1, // 10 FPS
		},
		Debug: DebugSettings{
			Enabled:                   true,
			FontPath:                  "assets/fonts/Mono_10", // Will append .sheet.png/.sheet.json
			FontScale:                 1.0,                    // 1:1 scale (no additional scaling beyond pixel scale)
			CharacterSpacingReduction: 8.0,                    // Reduce spacing by 8 pixels (adjust as needed)
			MaxMessages:               10,
			MessageLifetime:           0, // 0 = never fade (keep all messages)
			ConsoleHeight:             200.0,
			BackgroundColor:           [4]float32{0.0, 0.0, 0.0, 0.7}, // Semi-transparent black
			TextColor:                 [4]float32{0.0, 1.0, 0.0, 1.0}, // Green text (classic terminal look)
		},
		Rendering: RenderingSettings{
			PixelArtMode:        true,      // Enable pixel-perfect rendering
			TextureFiltering:    "nearest", // Use nearest-neighbor filtering for pixel art
			PixelPerfectScaling: true,      // Ensure integer scaling
			PixelScale:          5,         // 5 real pixels per game pixel (5x upscaling)
			UILineSpacing:       1.1,       // UI elements line spacing (menus, logs, status)
			TextLineSpacing:     1.1,       // Paragraph text line spacing (newlines in strings)
		},
	}
}

// WindowWidth returns the actual window pixel width for these settings: the
// virtual resolution width scaled up by the pixel scale.
func (s Settings) WindowWidth() int {
	return int(s.Screen.Width) * s.Rendering.PixelScale
}

// WindowHeight returns the actual window pixel height for these settings: the
// virtual resolution height scaled up by the pixel scale.
func (s Settings) WindowHeight() int {
	return int(s.Screen.Height) * s.Rendering.PixelScale
}

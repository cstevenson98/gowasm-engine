package config

// Settings contains all global configuration for the game engine
type Settings struct {
	Screen    ScreenSettings
	Player    PlayerSettings
	Animation AnimationSettings
	Debug     DebugSettings
	Rendering RenderingSettings
}

// ScreenSettings contains display and canvas configuration
type ScreenSettings struct {
	Width  float64
	Height float64
}

// PlayerSettings contains player-specific configuration
type PlayerSettings struct {
	SpawnX       float64 // X position for player spawn (0 = left, Screen.Width = right)
	SpawnY       float64 // Y position for player spawn (0 = top, Screen.Height = bottom)
	Size         float64 // Player sprite size (width and height)
	Speed        float64 // Movement speed in pixels per second
	TexturePath  string  // Path to player texture
	SpriteColumns int    // Number of columns in sprite sheet
	SpriteRows    int    // Number of rows in sprite sheet
}

// AnimationSettings contains animation timing configuration
type AnimationSettings struct {
	PlayerFrameTime  float64 // Time per frame for player animation (seconds)
	DefaultFrameTime float64 // Default frame time for other sprites (seconds)
}

// DebugSettings contains debug console configuration
type DebugSettings struct {
	Enabled                bool      // Enable/disable debug console
	FontPath               string    // Path to font sprite sheet (without .sheet.png extension)
	FontScale              float64   // Scale factor for debug text (1.0 = normal, 2.0 = double size)
	CharacterSpacingReduction float64 // Pixels to reduce character spacing (reduces padding between letters)
	MaxMessages            int       // Maximum number of messages to display
	MessageLifetime        float64   // Time before messages fade out (0 = never fade)
	ConsoleHeight          float64   // Height of the console in pixels
	BackgroundColor        [4]float32 // RGBA background color (with alpha for transparency)
	TextColor              [4]float32 // RGBA text color
}

// RenderingSettings contains rendering quality and style configuration
type RenderingSettings struct {
	PixelArtMode           bool    // Enable pixel-perfect rendering (nearest-neighbor filtering)
	TextureFiltering       string  // "nearest" or "linear" - texture filtering mode
	PixelPerfectScaling    bool    // Ensure integer scaling for pixel art
}

// Global is the global settings instance
var Global = Settings{
	Screen: ScreenSettings{
		Width:  800.0,
		Height: 600.0,
	},
	Player: PlayerSettings{
		SpawnX:       0.0, // Will be calculated as center in scene
		SpawnY:       0.0, // Will be calculated as center in scene
		Size:         128.0,
		Speed:        200.0, // pixels per second
		TexturePath:  "llama.png",
		SpriteColumns: 2,
		SpriteRows:    3,
	},
	Animation: AnimationSettings{
		PlayerFrameTime:  0.15, // 6.67 FPS
		DefaultFrameTime: 0.1,  // 10 FPS
	},
	Debug: DebugSettings{
		Enabled:                  true,
		FontPath:                "fonts/Mono_10", // Will append .sheet.png/.sheet.json
		FontScale:               1.5,              // Scale up for better readability
		CharacterSpacingReduction: 8.0,            // Reduce spacing by 8 pixels (adjust as needed)
		MaxMessages:             10,
		MessageLifetime:         0, // 0 = never fade (keep all messages)
		ConsoleHeight:           200.0,
		BackgroundColor:         [4]float32{0.0, 0.0, 0.0, 0.7}, // Semi-transparent black
		TextColor:               [4]float32{0.0, 1.0, 0.0, 1.0}, // Green text (classic terminal look)
	},
	Rendering: RenderingSettings{
		PixelArtMode:        true,  // Enable pixel-perfect rendering
		TextureFiltering:    "nearest", // Use nearest-neighbor filtering for pixel art
		PixelPerfectScaling: true,  // Ensure integer scaling
	},
}

// GetPlayerSpawnPosition calculates the centered spawn position for the player
func GetPlayerSpawnPosition() (x, y float64) {
	x = (Global.Screen.Width - Global.Player.Size) / 2
	y = (Global.Screen.Height - Global.Player.Size) / 2
	return
}


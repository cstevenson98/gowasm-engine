package config

// Settings contains all global configuration for the game engine
type Settings struct {
	Screen    ScreenSettings
	Player    PlayerSettings
	Animation AnimationSettings
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
	PlayerFrameTime float64 // Time per frame for player animation (seconds)
	DefaultFrameTime float64 // Default frame time for other sprites (seconds)
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
}

// GetPlayerSpawnPosition calculates the centered spawn position for the player
func GetPlayerSpawnPosition() (x, y float64) {
	x = (Global.Screen.Width - Global.Player.Size) / 2
	y = (Global.Screen.Height - Global.Player.Size) / 2
	return
}


package config

// Settings contains all global configuration for the game engine
type Settings struct {
	Screen    ScreenSettings
	Player    PlayerSettings
	Animation AnimationSettings
	Debug     DebugSettings
	Rendering RenderingSettings
	Battle    BattleSettings
}

// ScreenSettings contains display and canvas configuration
type ScreenSettings struct {
	Width        float64 // Virtual game resolution width
	Height       float64 // Virtual game resolution height
	CanvasWidth  int     // Actual canvas pixel width
	CanvasHeight int     // Actual canvas pixel height
}

// PlayerSettings contains player-specific configuration
type PlayerSettings struct {
	SpawnX        float64 // X position for player spawn (0 = left, Screen.Width = right)
	SpawnY        float64 // Y position for player spawn (0 = top, Screen.Height = bottom)
	Size          float64 // Player sprite size (width and height)
	Speed         float64 // Movement speed in pixels per second
	TexturePath   string  // Path to player texture
	SpriteColumns int     // Number of columns in sprite sheet
	SpriteRows    int     // Number of rows in sprite sheet
}

// AnimationSettings contains animation timing configuration
type AnimationSettings struct {
	PlayerFrameTime  float64 // Time per frame for player animation (seconds)
	DefaultFrameTime float64 // Default frame time for other sprites (seconds)
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

// BattleSettings contains battle scene configuration
type BattleSettings struct {
	PlayerHP      int     // Player's starting HP
	PlayerMaxHP   int     // Player's maximum HP
	EnemyHP       int     // Enemy's starting HP
	EnemyMaxHP    int     // Enemy's maximum HP
	EnemyTexture  string  // Path to enemy texture
	MenuFontPath  string  // Path to menu font (without .sheet.png extension)
	MenuFontScale float64 // Scale factor for menu text

	// Battle system configuration
	TimerChargeRate      float64 // How fast action timers charge (1.0 = 1.0 per second)
	AnimationDuration    float64 // Default animation duration in seconds
	DamageEffectDuration float64 // How long damage numbers are displayed
	ActionQueueSize      int     // Size of the action queue buffer
}

// Global is the global settings instance
var Global = Settings{
	Screen: ScreenSettings{
		Width:        800.0, // Virtual game resolution
		Height:       600.0, // Virtual game resolution
		CanvasWidth:  960,   // Actual canvas size (2x virtual for 4x pixel scale)
		CanvasHeight: 720,   // Actual canvas size (2x virtual for 4x pixel scale)
	},
	Player: PlayerSettings{
		SpawnX:        0.0,   // Will be calculated as center in scene
		SpawnY:        0.0,   // Will be calculated as center in scene
		Size:          32.0,  // Native sprite frame size (1:1 with texture, will be scaled by PixelScale)
		Speed:         200.0, // pixels per second
		TexturePath:   "llama.png",
		SpriteColumns: 2,
		SpriteRows:    3,
	},
	Animation: AnimationSettings{
		PlayerFrameTime:  0.15, // 6.67 FPS
		DefaultFrameTime: 0.1,  // 10 FPS
	},
	Debug: DebugSettings{
		Enabled:                   true,
		FontPath:                  "fonts/Mono_10", // Will append .sheet.png/.sheet.json
		FontScale:                 1.0,             // 1:1 scale (no additional scaling beyond pixel scale)
		CharacterSpacingReduction: 8.0,             // Reduce spacing by 8 pixels (adjust as needed)
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
		PixelScale:          3,         // 3 real pixels per game pixel (3x upscaling)
		UILineSpacing:       1.1,       // UI elements line spacing (menus, logs, status)
		TextLineSpacing:     1.1,       // Paragraph text line spacing (newlines in strings)
	},
	Battle: BattleSettings{
		PlayerHP:      100,
		PlayerMaxHP:   100,
		EnemyHP:       80,
		EnemyMaxHP:    80,
		EnemyTexture:  "art/ghost.png",
		MenuFontPath:  "fonts/Mono_10",
		MenuFontScale: 1.0,

		// Battle system configuration
		TimerChargeRate:      0.33, // 0.33 per second (3 seconds to fill)
		AnimationDuration:    1.0,  // 1 second default
		DamageEffectDuration: 2.0,  // 2 seconds for damage numbers
		ActionQueueSize:      100,  // Buffer for 100 actions
	},
}

// GetPlayerSpawnPosition calculates the centered spawn position for the player
func GetPlayerSpawnPosition() (x, y float64) {
	x = (Global.Screen.Width - Global.Player.Size) / 2
	y = (Global.Screen.Height - Global.Player.Size) / 2
	return
}

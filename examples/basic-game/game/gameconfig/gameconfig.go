// Package gameconfig holds configuration specific to the basic-game example:
// the player's appearance/stats and battle-mode content (enemy, timing,
// menus).
//
// Unlike the engine's github.com/cstevenson98/gowasm-engine/pkg/config.Settings
// (which has no global instance and is instead constructed explicitly and
// passed to engine.NewEngine), this package does expose a package-level
// global. That's deliberate: it configures one specific game, not a reusable
// engine that other games embed, so a singleton is the simplest fit and
// mirrors how the rest of this example already reads its own settings.
package gameconfig

// Settings contains this game's configuration.
type Settings struct {
	Player Player
	Battle Battle

	// PlayerFrameTime is the seconds-per-frame for the player's walk
	// animation.
	PlayerFrameTime float64
}

// Player contains the player character's appearance and movement.
type Player struct {
	Size          float64 // Sprite size (width and height), native 1:1 with the texture.
	Speed         float64 // Movement speed in pixels per second.
	TexturePath   string  // Path to the player texture.
	SpriteColumns int     // Number of columns in the sprite sheet.
	SpriteRows    int     // Number of rows in the sprite sheet.
}

// Battle contains battle-mode content and timing.
type Battle struct {
	PlayerHP      int
	PlayerMaxHP   int
	EnemyHP       int
	EnemyMaxHP    int
	EnemyTexture  string  // Path to enemy texture.
	MenuFontPath  string  // Path to menu font (without .sheet.png extension).
	MenuFontScale float64 // Scale factor for menu text.

	TimerChargeRate      float64 // How fast action timers charge (1.0 = 1.0 per second).
	AnimationDuration    float64 // Default animation duration in seconds.
	DamageEffectDuration float64 // How long damage numbers are displayed.
	ActionQueueSize      int     // Size of the action queue buffer.
}

// Global is this game's configuration singleton.
var Global = Settings{
	Player: Player{
		Size:          32.0, // Native sprite frame size (scaled by the engine's PixelScale).
		Speed:         20.0, // pixels per second
		TexturePath:   "assets/llama.png",
		SpriteColumns: 2,
		SpriteRows:    3,
	},
	PlayerFrameTime: 0.15, // 6.67 FPS
	Battle: Battle{
		PlayerHP:      100,
		PlayerMaxHP:   100,
		EnemyHP:       80,
		EnemyMaxHP:    80,
		EnemyTexture:  "assets/art/ghost.png",
		MenuFontPath:  "assets/fonts/Mono_10",
		MenuFontScale: 1.0,

		TimerChargeRate:      0.33, // 0.33 per second (3 seconds to fill)
		AnimationDuration:    1.0,  // 1 second default
		DamageEffectDuration: 2.0,  // 2 seconds for damage numbers
		ActionQueueSize:      100,  // Buffer for 100 actions
	},
}

// GetPlayerSpawnPosition calculates the centered spawn position for the
// player within a screenWidth x screenHeight virtual screen (the engine's
// config.Settings.Screen dimensions).
func GetPlayerSpawnPosition(screenWidth, screenHeight float64) (x, y float64) {
	x = (screenWidth - Global.Player.Size) / 2
	y = (screenHeight - Global.Player.Size) / 2
	return
}

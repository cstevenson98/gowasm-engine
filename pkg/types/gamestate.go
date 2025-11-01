package types

// GameState represents different states of the game
type GameState int

const (
	// MENU is the main menu state
	MENU GameState = iota
	// GAMEPLAY is the sprite rendering state
	GAMEPLAY
	// PLAYER_MENU is the player menu state (accessible from gameplay)
	PLAYER_MENU
	// TRIANGLE is the triangle rendering state
	TRIANGLE
	// BATTLE is the battle scene state
	BATTLE
)

// String returns the string representation of the game state
func (g GameState) String() string {
	switch g {
	case MENU:
		return "MENU"
	case GAMEPLAY:
		return "GAMEPLAY"
	case PLAYER_MENU:
		return "PLAYER_MENU"
	case TRIANGLE:
		return "TRIANGLE"
	case BATTLE:
		return "BATTLE"
	default:
		return "UNKNOWN"
	}
}

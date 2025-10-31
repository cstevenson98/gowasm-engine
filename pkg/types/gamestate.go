package types

// GameState represents different states of the game
type GameState int

const (
	// GAMEPLAY is the sprite rendering state
	GAMEPLAY GameState = iota
	// TRIANGLE is the triangle rendering state
	TRIANGLE
	// BATTLE is the battle scene state
	BATTLE
)

// String returns the string representation of the game state
func (g GameState) String() string {
	switch g {
	case GAMEPLAY:
		return "GAMEPLAY"
	case TRIANGLE:
		return "TRIANGLE"
	case BATTLE:
		return "BATTLE"
	default:
		return "UNKNOWN"
	}
}

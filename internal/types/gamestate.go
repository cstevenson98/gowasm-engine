package types

// GameState represents different states of the game
type GameState int

const (
	// GAMEPLAY is the sprite rendering state
	GAMEPLAY GameState = iota
	// TRIANGLE is the triangle rendering state
	TRIANGLE
)

// String returns the string representation of the game state
func (g GameState) String() string {
	switch g {
	case GAMEPLAY:
		return "SPRITE"
	case TRIANGLE:
		return "TRIANGLE"
	default:
		return "UNKNOWN"
	}
}

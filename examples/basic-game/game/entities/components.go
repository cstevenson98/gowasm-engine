// Package entities defines the game's ECS components and entity spawners. In the
// ECS model there is no Player or Enemy "object": a player is an entity carrying
// a PlayerControl component (plus Position, Sprite, ...); behaviour lives in
// systems. Battle participants are represented by a small adapter (see
// battle.go) that implements the battle system's BattleEntity interface.
package entities

// PlayerControl marks the player-controlled entity and carries its movement
// speed. The PlayerInputSystem reads this to convert input into velocity.
type PlayerControl struct {
	Speed float64
}

// Stats holds the player's persistent RPG statistics (mirrors gamestate.PlayerStats).
type Stats struct {
	Level      int
	HP         int
	MaxHP      int
	Experience int
}

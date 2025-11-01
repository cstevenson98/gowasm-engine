//go:build js

package gamestate

import (
	"time"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GameState represents the complete game state that can be saved and loaded
type GameState struct {
	Version        int             `json:"version"`   // Version for migration support
	Timestamp      int64           `json:"timestamp"` // Unix timestamp in milliseconds
	PlayerStats    PlayerStats     `json:"player_stats"`
	PlayerPosition types.Vector2   `json:"player_position"`
	StoryState     map[string]bool `json:"story_state"` // Quest flags, NPC states, etc.
}

// PlayerStats represents player character statistics
type PlayerStats struct {
	Level      int `json:"level"`
	HP         int `json:"hp"`
	MaxHP      int `json:"max_hp"`
	Experience int `json:"experience"`
	// Add more stats as needed
}

// SaveInfo represents metadata about a save slot
type SaveInfo struct {
	Key         string `json:"key"`          // localStorage key
	Timestamp   int64  `json:"timestamp"`    // Unix timestamp in milliseconds
	DisplayTime string `json:"display_time"` // Human-readable time
	PlayerLevel int    `json:"player_level"`
	PlayerHP    int    `json:"player_hp"`
	PlayerMaxHP int    `json:"player_max_hp"`
}

// Current version of save format
const GameStateVersion = 1

// FormatTimestamp formats a Unix timestamp (ms) into a human-readable string
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp/1000, (timestamp%1000)*1000000)
	return t.Format("2006-01-02 15:04:05")
}

//go:build js

package gamestate

import (
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GameStateManager manages the game's global state with save/load functionality
type GameStateManager struct {
	currentState *GameState
	playerRef    interface{} // Game-specific player reference (type defined by game, not engine)
	mu           sync.Mutex
}

// NewGameStateManager creates a new game state manager
func NewGameStateManager() *GameStateManager {
	return &GameStateManager{
		currentState: nil,
	}
}

// GetState returns the current game state (thread-safe)
func (gsm *GameStateManager) GetState() *GameState {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	return gsm.currentState
}

// SetState sets the current game state (thread-safe)
func (gsm *GameStateManager) SetState(state *GameState) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	gsm.currentState = state
}

// SetPlayer sets the current player reference (thread-safe)
// The player is part of the game state, stored separately for easy access
func (gsm *GameStateManager) SetPlayer(player interface{}) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	gsm.playerRef = player
	logger.Logger.Debugf("Updated player reference in game state manager")
}

// GetPlayer returns the current player reference (thread-safe)
// Returns nil if no player is set
func (gsm *GameStateManager) GetPlayer() interface{} {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	return gsm.playerRef
}

// CreateNewGame initializes a new game with default state
func (gsm *GameStateManager) CreateNewGame() error {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	spawnX, spawnY := config.GetPlayerSpawnPosition()

	gsm.currentState = &GameState{
		Version:   GameStateVersion,
		Timestamp: 0, // Set when saving
		PlayerStats: PlayerStats{
			Level:      1,
			HP:         config.Global.Battle.PlayerHP,
			MaxHP:      config.Global.Battle.PlayerMaxHP,
			Experience: 0,
		},
		PlayerPosition: types.Vector2{
			X: spawnX,
			Y: spawnY,
		},
		StoryState: make(map[string]bool),
	}

	logger.Logger.Debugf("Created new game state")
	return nil
}

// UpdateStateFromPlayer updates the game state with current player position and stats
// This should be called before saving to ensure the state is up-to-date
func (gsm *GameStateManager) UpdateStateFromPlayer(playerPosition types.Vector2, playerStats PlayerStats) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	if gsm.currentState == nil {
		return
	}

	gsm.currentState.PlayerPosition = playerPosition
	gsm.currentState.PlayerStats = playerStats
	logger.Logger.Debugf("Updated game state - Position: (%.2f, %.2f), HP: %d/%d", playerPosition.X, playerPosition.Y, playerStats.HP, playerStats.MaxHP)
}

// SaveCurrentGame saves the current game state to localStorage
// Returns the save key (timestamp-based) and any error
func (gsm *GameStateManager) SaveCurrentGame() (string, error) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	if gsm.currentState == nil {
		return "", fmt.Errorf("no game state to save")
	}

	// Set timestamp to current time
	gsm.currentState.Timestamp = int64(js.Global().Get("Date").New().Call("getTime").Float())

	// Serialize to JSON
	data, err := json.Marshal(gsm.currentState)
	if err != nil {
		return "", fmt.Errorf("failed to marshal game state: %w", err)
	}

	// Create save key from timestamp
	saveKey := fmt.Sprintf("%s%d", localStorageKeyPrefix, gsm.currentState.Timestamp)

	// Save to localStorage
	err = SaveToLocalStorage(saveKey, data)
	if err != nil {
		return "", fmt.Errorf("failed to save to localStorage: %w", err)
	}

	// Update save index
	err = gsm.updateSaveIndex(saveKey, gsm.currentState)
	if err != nil {
		logger.Logger.Warnf("Failed to update save index: %s", err.Error())
		// Don't fail the save if index update fails
	}

	logger.Logger.Infof("Saved game state: %s", saveKey)
	return saveKey, nil
}

// LoadSave loads a specific save from localStorage by key
func (gsm *GameStateManager) LoadSave(saveKey string) error {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	// Load from localStorage
	data, err := LoadFromLocalStorage(saveKey)
	if err != nil {
		return fmt.Errorf("failed to load from localStorage: %w", err)
	}

	// Deserialize from JSON
	var state GameState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	// Validate version (future: handle migration)
	if state.Version != GameStateVersion {
		logger.Logger.Warnf("Save version mismatch: got %d, expected %d", state.Version, GameStateVersion)
	}

	gsm.currentState = &state
	logger.Logger.Infof("Loaded game state: %s", saveKey)
	return nil
}

// ListSaves returns a list of all available saves with metadata
func (gsm *GameStateManager) ListSaves() ([]SaveInfo, error) {
	// Load index first (faster)
	indexSaves, err := LoadSaveIndex()
	if err == nil && len(indexSaves) > 0 {
		return indexSaves, nil
	}

	// Fallback: scan localStorage keys
	keys := ListKeys(localStorageKeyPrefix)
	var saves []SaveInfo

	for _, key := range keys {
		data, err := LoadFromLocalStorage(key)
		if err != nil {
			logger.Logger.Warnf("Failed to load save %s: %s", key, err.Error())
			continue
		}

		var state GameState
		err = json.Unmarshal(data, &state)
		if err != nil {
			logger.Logger.Warnf("Failed to unmarshal save %s: %s", key, err.Error())
			continue
		}

		saves = append(saves, SaveInfo{
			Key:         key,
			Timestamp:   state.Timestamp,
			DisplayTime: FormatTimestamp(state.Timestamp),
			PlayerLevel: state.PlayerStats.Level,
			PlayerHP:    state.PlayerStats.HP,
			PlayerMaxHP: state.PlayerStats.MaxHP,
		})
	}

	// Sort by timestamp (newest first) - simple insertion sort
	for i := 1; i < len(saves); i++ {
		for j := i; j > 0 && saves[j].Timestamp > saves[j-1].Timestamp; j-- {
			saves[j], saves[j-1] = saves[j-1], saves[j]
		}
	}

	return saves, nil
}

// DeleteSave deletes a save from localStorage
func (gsm *GameStateManager) DeleteSave(saveKey string) error {
	err := DeleteFromLocalStorage(saveKey)
	if err != nil {
		return fmt.Errorf("failed to delete save: %w", err)
	}

	// Update save index
	err = gsm.rebuildSaveIndex()
	if err != nil {
		logger.Logger.Warnf("Failed to rebuild save index: %s", err.Error())
	}

	logger.Logger.Infof("Deleted save: %s", saveKey)
	return nil
}

// updateSaveIndex updates the save index with a new save
func (gsm *GameStateManager) updateSaveIndex(saveKey string, state *GameState) error {
	saves, _ := LoadSaveIndex()

	// Check if save already exists in index
	found := false
	for i, save := range saves {
		if save.Key == saveKey {
			// Update existing entry
			saves[i] = SaveInfo{
				Key:         saveKey,
				Timestamp:   state.Timestamp,
				DisplayTime: FormatTimestamp(state.Timestamp),
				PlayerLevel: state.PlayerStats.Level,
				PlayerHP:    state.PlayerStats.HP,
				PlayerMaxHP: state.PlayerStats.MaxHP,
			}
			found = true
			break
		}
	}

	if !found {
		// Add new entry
		saves = append(saves, SaveInfo{
			Key:         saveKey,
			Timestamp:   state.Timestamp,
			DisplayTime: FormatTimestamp(state.Timestamp),
			PlayerLevel: state.PlayerStats.Level,
			PlayerHP:    state.PlayerStats.HP,
			PlayerMaxHP: state.PlayerStats.MaxHP,
		})
	}

	// Sort by timestamp (newest first)
	for i := 1; i < len(saves); i++ {
		for j := i; j > 0 && saves[j].Timestamp > saves[j-1].Timestamp; j-- {
			saves[j], saves[j-1] = saves[j-1], saves[j]
		}
	}

	return SaveSaveIndex(saves)
}

// rebuildSaveIndex rebuilds the save index by scanning all save keys
func (gsm *GameStateManager) rebuildSaveIndex() error {
	saves, err := gsm.ListSaves()
	if err != nil {
		return err
	}

	return SaveSaveIndex(saves)
}

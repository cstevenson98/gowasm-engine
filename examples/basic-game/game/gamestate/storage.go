package gamestate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// Storage directory for game saves
const saveDir = ".gowasm-game-saves"

// localStorageKeyPrefix is the prefix for all game save keys
const localStorageKeyPrefix = "game_save_"
const localStorageIndexKey = "game_saves_index"

// getSaveDir returns the full path to the save directory
func getSaveDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	savePath := filepath.Join(homeDir, saveDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %w", err)
	}

	return savePath, nil
}

// keyToFilename converts a storage key to a safe filename
func keyToFilename(key string) string {
	// Replace any unsafe characters
	safe := strings.ReplaceAll(key, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	return safe + ".json"
}

// SaveToLocalStorage saves data to a file (desktop implementation)
func SaveToLocalStorage(key string, data []byte) error {
	savePath, err := getSaveDir()
	if err != nil {
		return err
	}

	filename := filepath.Join(savePath, keyToFilename(key))

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	logger.Logger.Debugf("Saved to file: %s (%d bytes)", filename, len(data))
	return nil
}

// LoadFromLocalStorage loads data from a file (desktop implementation)
func LoadFromLocalStorage(key string) ([]byte, error) {
	savePath, err := getSaveDir()
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(savePath, keyToFilename(key))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	logger.Logger.Debugf("Loaded from file: %s (%d bytes)", filename, len(data))
	return data, nil
}

// DeleteFromLocalStorage deletes a file (desktop implementation)
func DeleteFromLocalStorage(key string) error {
	savePath, err := getSaveDir()
	if err != nil {
		return err
	}

	filename := filepath.Join(savePath, keyToFilename(key))

	err = os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	logger.Logger.Debugf("Deleted file: %s", filename)
	return nil
}

// ListKeys returns all storage keys with the given prefix (desktop implementation)
func ListKeys(prefix string) []string {
	savePath, err := getSaveDir()
	if err != nil {
		logger.Logger.Warnf("Failed to get save directory: %s", err)
		return []string{}
	}

	entries, err := os.ReadDir(savePath)
	if err != nil {
		logger.Logger.Warnf("Failed to read save directory: %s", err)
		return []string{}
	}

	var keys []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Remove .json extension and check prefix
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			key := strings.TrimSuffix(name, ".json")
			if strings.HasPrefix(key, prefix) {
				keys = append(keys, key)
			}
		}
	}

	return keys
}

// LoadSaveIndex loads the save index from storage
func LoadSaveIndex() ([]SaveInfo, error) {
	data, err := LoadFromLocalStorage(localStorageIndexKey)
	if err != nil {
		// No index yet - return empty list
		return []SaveInfo{}, nil
	}

	var saves []SaveInfo
	err = json.Unmarshal(data, &saves)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal save index: %w", err)
	}

	return saves, nil
}

// SaveSaveIndex saves the save index to storage
func SaveSaveIndex(saves []SaveInfo) error {
	data, err := json.Marshal(saves)
	if err != nil {
		return fmt.Errorf("failed to marshal save index: %w", err)
	}

	return SaveToLocalStorage(localStorageIndexKey, data)
}

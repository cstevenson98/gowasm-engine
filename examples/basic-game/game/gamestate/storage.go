//go:build js

package gamestate

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// localStorageKeyPrefix is the prefix for all game save keys
const localStorageKeyPrefix = "game_save_"
const localStorageIndexKey = "game_saves_index"

// SaveToLocalStorage saves data to browser localStorage
func SaveToLocalStorage(key string, data []byte) error {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return fmt.Errorf("localStorage not available")
	}

	// Convert byte slice to base64 string using btoa
	// btoa expects a string where each byte is a character (Latin-1 encoding)
	// Convert bytes to string and wrap in js.ValueOf for proper JavaScript conversion
	dataString := string(data)
	encodedStr := js.Global().Call("btoa", js.ValueOf(dataString)).String()

	// Save to localStorage
	localStorage.Call("setItem", key, encodedStr)
	logger.Logger.Debugf("Saved to localStorage: %s (%d bytes)", key, len(data))
	return nil
}

// LoadFromLocalStorage loads data from browser localStorage
func LoadFromLocalStorage(key string) ([]byte, error) {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return nil, fmt.Errorf("localStorage not available")
	}

	// Get from localStorage
	dataStr := localStorage.Call("getItem", key)
	if dataStr.IsNull() || dataStr.IsUndefined() {
		return nil, fmt.Errorf("key not found in localStorage: %s", key)
	}

	// Decode base64 string back to bytes using atob global function
	decoded := js.Global().Call("atob", dataStr).String()
	logger.Logger.Debugf("Loaded from localStorage: %s (%d bytes)", key, len(decoded))
	return []byte(decoded), nil
}

// DeleteFromLocalStorage deletes a key from browser localStorage
func DeleteFromLocalStorage(key string) error {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return fmt.Errorf("localStorage not available")
	}

	localStorage.Call("removeItem", key)
	logger.Logger.Debugf("Deleted from localStorage: %s", key)
	return nil
}

// ListKeys returns all localStorage keys with the given prefix
func ListKeys(prefix string) []string {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return []string{}
	}

	var keys []string
	length := localStorage.Get("length").Int()

	for i := 0; i < length; i++ {
		key := localStorage.Call("key", i).String()
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}

	return keys
}

// LoadSaveIndex loads the save index from localStorage
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

// SaveSaveIndex saves the save index to localStorage
func SaveSaveIndex(saves []SaveInfo) error {
	data, err := json.Marshal(saves)
	if err != nil {
		return fmt.Errorf("failed to marshal save index: %w", err)
	}

	return SaveToLocalStorage(localStorageIndexKey, data)
}

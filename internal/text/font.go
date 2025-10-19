//go:build js

package text

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/types"
)

// CharacterData represents metadata for a single character in the font sprite sheet
type CharacterData struct {
	Index int     `json:"index"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	U0    float64 `json:"u0"`
	V0    float64 `json:"v0"`
	U1    float64 `json:"u1"`
	V1    float64 `json:"v1"`
}

// FontMetadata represents the JSON metadata from the font sprite sheet generator
type FontMetadata struct {
	FontName       string                   `json:"font_name"`
	FontSize       int                      `json:"font_size"`
	CellWidth      int                      `json:"cell_width"`
	CellHeight     int                      `json:"cell_height"`
	Columns        int                      `json:"columns"`
	Rows           int                      `json:"rows"`
	ImageWidth     int                      `json:"image_width"`
	ImageHeight    int                      `json:"image_height"`
	CharacterCount int                      `json:"character_count"`
	CharacterMap   map[string]CharacterData `json:"character_map"`
}

// SpriteFont implements the Font interface using a sprite sheet
type SpriteFont struct {
	texturePath string
	metadata    *FontMetadata
	loaded      bool
}

// NewSpriteFont creates a new SpriteFont instance
func NewSpriteFont() *SpriteFont {
	return &SpriteFont{
		loaded: false,
	}
}

// LoadFont loads a font sprite sheet from the given base path
// It expects both a .sheet.png and .sheet.json file
func (f *SpriteFont) LoadFont(basePath string) error {
	logger.Logger.Debugf("Loading font from: %s", basePath)

	// Store the texture path (PNG)
	f.texturePath = basePath + ".sheet.png"
	metadataPath := basePath + ".sheet.json"

	// Load the JSON metadata
	err := f.loadMetadata(metadataPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load font metadata: %s", err)
		return fmt.Errorf("failed to load font metadata: %w", err)
	}

	f.loaded = true
	logger.Logger.Debugf("Font loaded successfully: %s (%dx%d cells, %d characters)",
		f.metadata.FontName, f.metadata.CellWidth, f.metadata.CellHeight, f.metadata.CharacterCount)

	return nil
}

// loadMetadata loads the JSON metadata file
func (f *SpriteFont) loadMetadata(path string) error {
	// Use fetch API to load the JSON file
	promise := js.Global().Call("fetch", path)

	// Create channels for async handling
	done := make(chan error, 1)

	// Handle the promise
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		
		// Check if response is OK
		if !response.Get("ok").Bool() {
			done <- fmt.Errorf("failed to fetch metadata: HTTP %d", response.Get("status").Int())
			return nil
		}

		// Get JSON from response
		jsonPromise := response.Call("json")
		jsonPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			jsonData := args[0]

			// Convert to JSON string
			jsonString := js.Global().Get("JSON").Call("stringify", jsonData).String()

			// Parse JSON into our struct
			var metadata FontMetadata
			err := json.Unmarshal([]byte(jsonString), &metadata)
			if err != nil {
				done <- fmt.Errorf("failed to parse font metadata JSON: %w", err)
				return nil
			}

			f.metadata = &metadata
			done <- nil
			return nil
		}))

		jsonPromise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			done <- fmt.Errorf("failed to parse JSON response")
			return nil
		}))

		return nil
	}))

	promise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		done <- fmt.Errorf("fetch failed: %s", args[0].String())
		return nil
	}))

	// Wait for async operation to complete
	err := <-done
	return err
}

// GetCharacterUV returns the UV coordinates for a given character
func (f *SpriteFont) GetCharacterUV(char rune) (types.UVRect, error) {
	if !f.loaded {
		return types.UVRect{}, fmt.Errorf("font not loaded")
	}

	charStr := string(char)
	charData, exists := f.metadata.CharacterMap[charStr]
	if !exists {
		// Return UV for '?' as fallback if it exists, otherwise return error
		charData, exists = f.metadata.CharacterMap["?"]
		if !exists {
			return types.UVRect{}, fmt.Errorf("character not found: %c", char)
		}
		logger.Logger.Tracef("Character '%c' not found, using '?' as fallback", char)
	}

	// Convert from U0,V0,U1,V1 format to U,V,W,H format
	return types.UVRect{
		U: charData.U0,
		V: charData.V0,
		W: charData.U1 - charData.U0,
		H: charData.V1 - charData.V0,
	}, nil
}

// GetTexturePath returns the path to the font sprite sheet texture
func (f *SpriteFont) GetTexturePath() string {
	return f.texturePath
}

// GetCellSize returns the width and height of each character cell
func (f *SpriteFont) GetCellSize() (int, int) {
	if !f.loaded || f.metadata == nil {
		return 0, 0
	}
	return f.metadata.CellWidth, f.metadata.CellHeight
}

// IsLoaded returns true if the font is loaded and ready to use
func (f *SpriteFont) IsLoaded() bool {
	return f.loaded && f.metadata != nil
}


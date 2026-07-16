//go:build !js

package text

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

var (
	// fontCache stores loaded font metadata by base path to avoid reloading
	fontCache     = make(map[string]*FontMetadata)
	fontCacheLock sync.Mutex
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

// LoadFont loads a font sprite sheet from the given base path (desktop version using os.ReadFile)
// It expects both a .sheet.png and .sheet.json file
// Uses a global cache to avoid reloading the same font metadata multiple times
func (f *SpriteFont) LoadFont(basePath string) error {
	// Store the texture path (PNG)
	f.texturePath = basePath + ".sheet.png"
	metadataPath := basePath + ".sheet.json"

	// Check cache first
	fontCacheLock.Lock()
	cachedMetadata, exists := fontCache[basePath]
	fontCacheLock.Unlock()

	if exists {
		// Reuse cached metadata
		f.metadata = cachedMetadata
		f.loaded = true
		logger.Logger.Debugf("Using cached font: %s (%dx%d cells, %d characters)",
			f.metadata.FontName, f.metadata.CellWidth, f.metadata.CellHeight, f.metadata.CharacterCount)
		return nil
	}

	// Not in cache, load it
	logger.Logger.Debugf("Loading font from: %s", basePath)
	err := f.loadMetadata(metadataPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load font metadata: %s", err)
		return fmt.Errorf("failed to load font metadata: %w", err)
	}

	// Store in cache for future use
	fontCacheLock.Lock()
	fontCache[basePath] = f.metadata
	fontCacheLock.Unlock()

	f.loaded = true
	logger.Logger.Debugf("Font loaded and cached: %s (%dx%d cells, %d characters)",
		f.metadata.FontName, f.metadata.CellWidth, f.metadata.CellHeight, f.metadata.CharacterCount)

	return nil
}

// loadMetadata loads the JSON metadata file (desktop version using os.ReadFile)
func (f *SpriteFont) loadMetadata(path string) error {
	// Read file from filesystem
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read font metadata file: %w", err)
	}

	// Parse JSON into our struct
	var metadata FontMetadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return fmt.Errorf("failed to parse font metadata JSON: %w", err)
	}

	f.metadata = &metadata
	return nil
}

// GetTexturePath returns the path to the texture PNG file
func (f *SpriteFont) GetTexturePath() string {
	return f.texturePath
}

// GetCharacterUV returns the UV coordinates for a given character (Font interface method)
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
	}

	// Convert from U0,V0,U1,V1 format to U,V,W,H format
	return types.UVRect{
		U: charData.U0,
		V: charData.V0,
		W: charData.U1 - charData.U0,
		H: charData.V1 - charData.V0,
	}, nil
}

// GetCharacter returns the UV coordinates and dimensions for a character
func (f *SpriteFont) GetCharacter(char rune) (types.UVRect, types.Vector2, error) {
	if !f.loaded {
		return types.UVRect{}, types.Vector2{}, fmt.Errorf("font not loaded")
	}

	charData, exists := f.metadata.CharacterMap[string(char)]
	if !exists {
		// Return a default character (space) or error
		charData, exists = f.metadata.CharacterMap[" "]
		if !exists {
			return types.UVRect{}, types.Vector2{}, fmt.Errorf("character not found in font: %c", char)
		}
	}

	uv := types.UVRect{
		U: charData.U0,
		V: charData.V0,
		W: charData.U1 - charData.U0,
		H: charData.V1 - charData.V0,
	}

	size := types.Vector2{
		X: float64(f.metadata.CellWidth),
		Y: float64(f.metadata.CellHeight),
	}

	return uv, size, nil
}

// GetMetadata returns the font metadata
func (f *SpriteFont) GetMetadata() (*FontMetadata, error) {
	if !f.loaded {
		return nil, fmt.Errorf("font not loaded")
	}
	return f.metadata, nil
}

// IsLoaded returns whether the font is loaded
func (f *SpriteFont) IsLoaded() bool {
	return f.loaded
}

// GetCellSize returns the width and height of each character cell
func (f *SpriteFont) GetCellSize() (int, int) {
	if !f.loaded || f.metadata == nil {
		return 0, 0
	}
	return f.metadata.CellWidth, f.metadata.CellHeight
}

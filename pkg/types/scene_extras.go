package types

// SceneOverlayRenderer is an optional interface a Scene can implement
// to render additional overlays (menus, HUD, debug) each frame.
type SceneOverlayRenderer interface {
	// RenderOverlays should render any overlays for the scene.
	RenderOverlays() error
}

// SceneTextureProvider is an optional interface a Scene can implement
// to provide extra texture paths that should be preloaded by the engine
// (e.g., fonts used by overlays/menus).
type SceneTextureProvider interface {
	// GetExtraTexturePaths returns additional texture paths to preload.
	GetExtraTexturePaths() []string
}

// SceneInputProvider is an optional interface a Scene can implement
// to receive the engine's input capturer during initialization.
// The engine automatically injects the input capturer when scenes implement this interface.
type SceneInputProvider interface {
	// SetInputCapturer sets the input capturer for the scene.
	// Called by the engine during scene initialization.
	SetInputCapturer(inputCapturer InputCapturer)
}

// SceneStateChangeRequester is an optional interface a Scene can implement
// to request game state changes back to the engine.
// The engine automatically injects a state change callback when scenes implement this interface.
type SceneStateChangeRequester interface {
	// SetStateChangeCallback sets a callback function that the scene can call
	// to request a state change. Called by the engine during scene initialization.
	SetStateChangeCallback(callback func(state GameState) error)
}

// SceneAssets represents all assets required by a scene
type SceneAssets struct {
	// TexturePaths are paths to texture files (.png, etc.) that need to be loaded
	TexturePaths []string
	// FontPaths are paths to font sprite sheets (base path, engine will append .sheet.json)
	FontPaths []string
}

// SceneAssetProvider is an optional interface a Scene can implement
// to declare all assets it requires. The engine will preload these assets
// BEFORE calling Initialize() to prevent deadlocks from blocking I/O operations.
type SceneAssetProvider interface {
	// GetRequiredAssets returns all assets that must be loaded before the scene initializes.
	// This is called before Initialize() to preload everything synchronously.
	GetRequiredAssets() SceneAssets
}

// SceneStateful is an optional interface a Scene can implement
// to persist state between scene changes. When a scene implements this interface,
// the engine will call SaveState() before Cleanup() and RestoreState() after Initialize()
// to maintain scene state across switches.
type SceneStateful interface {
	// SaveState saves the current scene state before cleanup.
	// This is called by the engine before Cleanup() when switching away from the scene.
	SaveState()

	// RestoreState restores the previously saved scene state after initialization.
	// This is called by the engine after Initialize() when switching back to the scene.
	// If no state was previously saved, this should restore to default values.
	RestoreState()
}

// SceneGameStateUser is an optional interface a Scene can implement
// to receive access to the game's global state manager during initialization.
// The engine will inject the game state provider (registered by the game) into scenes
// that implement this interface. The engine does not know or care about the specific
// type of the game state - it just passes through whatever the game registers.
type SceneGameStateUser interface {
	// SetGameState sets the game state provider (manager).
	// Called by the engine during scene initialization.
	// The provider type is defined by the game, not the engine.
	SetGameState(gameState interface{})
}

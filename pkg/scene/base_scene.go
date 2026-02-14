//go:build js

package scene

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BaseScene provides default implementations for the Scene interface and all optional scene interfaces.
// Embed this struct in your custom scenes to avoid boilerplate code.
//
// This automatically implements:
//   - Scene (core interface)
//   - SceneInputProvider
//   - SceneChangeRequester
//   - SceneGameStateUser
//   - SceneAssetProvider
//   - SceneStateful
//   - SceneOverlayRenderer
//   - SceneTextureProvider
//
// Example usage:
//
//	type MyScene struct {
//	    *scene.BaseScene
//	    customField int
//	}
//
//	func NewMyScene(width, height float64) *MyScene {
//	    return &MyScene{
//	        BaseScene: scene.NewBaseScene("MyScene", width, height),
//	        customField: 42,
//	    }
//	}
//
//	// Override only what you need:
//	func (s *MyScene) Initialize() error {
//	    if err := s.BaseScene.Initialize(); err != nil {
//	        return err
//	    }
//	    // Your custom initialization
//	    return nil
//	}
type BaseScene struct {
	// Core scene properties
	name         string
	screenWidth  float64
	screenHeight float64

	// Layer management (thread-safe)
	layers     map[SceneLayer][]types.GameObject
	layerMutex sync.RWMutex

	// Optional interface fields (injected by engine)
	inputCapturer       types.InputCapturer
	stateChangeCallback func(state types.GameState) error
	canvasManager       canvas.CanvasManager
	gameStateManager    interface{}

	// Saved state (for SceneStateful)
	savedState map[string]interface{}

	// Required assets (for SceneAssetProvider)
	requiredAssets types.SceneAssets

	// Debug console (engine feature, auto-initialized)
	debugFont               text.Font
	debugTextRenderer       text.TextRenderer
	debugConsoleInitialized bool
}

// NewBaseScene creates a new BaseScene with the given name and screen dimensions.
// This initializes all internal structures but does not call Initialize().
func NewBaseScene(name string, width, height float64) *BaseScene {
	return &BaseScene{
		name:         name,
		screenWidth:  width,
		screenHeight: height,
		layers:       make(map[SceneLayer][]types.GameObject),
		savedState:   make(map[string]interface{}),
		requiredAssets: types.SceneAssets{
			TexturePaths: []string{},
			FontPaths:    []string{},
		},
	}
}

// ===== Core Scene Interface =====

// GetName returns the scene identifier.
// Implements Scene interface.
func (b *BaseScene) GetName() string {
	return b.name
}

// Initialize sets up the scene and initializes empty layers.
// Override this method to add custom initialization logic.
// Remember to call b.BaseScene.Initialize() first in your override.
// Implements Scene interface.
func (b *BaseScene) Initialize() error {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()

	// Initialize empty layers if not already initialized
	if b.layers == nil {
		b.layers = make(map[SceneLayer][]types.GameObject)
	}

	// Ensure all standard layers exist
	if _, exists := b.layers[BACKGROUND]; !exists {
		b.layers[BACKGROUND] = []types.GameObject{}
	}
	if _, exists := b.layers[ENTITIES]; !exists {
		b.layers[ENTITIES] = []types.GameObject{}
	}
	if _, exists := b.layers[UI]; !exists {
		b.layers[UI] = []types.GameObject{}
	}

	return nil
}

// Update updates all game objects in all layers.
// Override this method to add custom update logic.
// You can call b.BaseScene.Update(deltaTime) to update all objects,
// then add your custom logic.
// Implements Scene interface.
func (b *BaseScene) Update(deltaTime float64) {
	b.layerMutex.RLock()
	defer b.layerMutex.RUnlock()

	// Update all game objects in all layers
	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
		if objects, exists := b.layers[layer]; exists {
			for _, obj := range objects {
				obj.Update(deltaTime)
			}
		}
	}
}

// GetRenderables returns all game objects in the correct render order (layer order).
// BACKGROUND is rendered first (back), then ENTITIES (middle), then UI (front).
// Implements Scene interface.
func (b *BaseScene) GetRenderables() []types.GameObject {
	b.layerMutex.RLock()
	defer b.layerMutex.RUnlock()

	renderables := []types.GameObject{}

	// Concatenate in layer order: BACKGROUND -> ENTITIES -> UI
	renderables = append(renderables, b.layers[BACKGROUND]...)
	renderables = append(renderables, b.layers[ENTITIES]...)
	renderables = append(renderables, b.layers[UI]...)

	return renderables
}

// Cleanup releases scene resources and clears all layers.
// Override this method to add custom cleanup logic.
// Remember to call b.BaseScene.Cleanup() in your override.
// Implements Scene interface.
func (b *BaseScene) Cleanup() {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()

	// Clear all layers
	for layer := range b.layers {
		b.layers[layer] = nil
	}
}

// ===== SceneInputProvider Interface =====

// SetInputCapturer sets the input capturer for the scene.
// Called automatically by the engine during scene initialization.
// Implements SceneInputProvider interface.
func (b *BaseScene) SetInputCapturer(inputCapturer types.InputCapturer) {
	b.inputCapturer = inputCapturer
}

// GetInputState returns the current input state.
// Returns an empty InputState if no input capturer is set.
func (b *BaseScene) GetInputState() types.InputState {
	if b.inputCapturer != nil {
		return b.inputCapturer.GetInputState()
	}
	return types.InputState{}
}

// ===== SceneChangeRequester Interface =====

// SetStateChangeCallback sets a callback function that the scene can call
// to request a pipeline state change.
// Called automatically by the engine during scene initialization.
// Implements SceneChangeRequester interface.
func (b *BaseScene) SetStateChangeCallback(callback func(state types.GameState) error) {
	b.stateChangeCallback = callback
}

// RequestStateChange requests a change to a different pipeline state (scene).
// Returns an error if the state change callback is not set or the state change fails.
func (b *BaseScene) RequestStateChange(newState types.GameState) error {
	if b.stateChangeCallback != nil {
		return b.stateChangeCallback(newState)
	}
	return nil // No-op if callback not set
}

// ===== SceneGameStateUser Interface =====

// SetGameState sets the game state provider (manager).
// Called automatically by the engine during scene initialization.
// The provider type is defined by the game, not the engine.
// Implements SceneGameStateUser interface.
func (b *BaseScene) SetGameState(gameState interface{}) {
	b.gameStateManager = gameState
}

// GetGameState returns the game state provider.
// Returns nil if no game state is set.
func (b *BaseScene) GetGameState() interface{} {
	return b.gameStateManager
}

// ===== SceneTextureProvider Interface =====

// SetCanvasManager sets the canvas manager for rendering.
// Called automatically during scene initialization (via dependency injection).
func (b *BaseScene) SetCanvasManager(cm canvas.CanvasManager) {
	b.canvasManager = cm
}

// GetCanvasManager returns the canvas manager.
// Returns nil if no canvas manager is set.
func (b *BaseScene) GetCanvasManager() canvas.CanvasManager {
	return b.canvasManager
}

// GetExtraTexturePaths returns additional texture paths to preload.
// Default implementation returns an empty slice.
// Override this if you need to preload extra textures.
// Implements SceneTextureProvider interface.
func (b *BaseScene) GetExtraTexturePaths() []string {
	return []string{}
}

// ===== SceneAssetProvider Interface =====

// GetRequiredAssets returns all assets that must be loaded before the scene initializes.
// This is called before Initialize() to preload everything synchronously.
// Implements SceneAssetProvider interface.
func (b *BaseScene) GetRequiredAssets() types.SceneAssets {
	return b.requiredAssets
}

// SetRequiredAssets sets the assets required by this scene.
// Call this before the scene is registered with the engine.
func (b *BaseScene) SetRequiredAssets(assets types.SceneAssets) {
	b.requiredAssets = assets
}

// AddRequiredTexture adds a texture path to the required assets.
// Convenience method for adding textures one at a time.
func (b *BaseScene) AddRequiredTexture(texturePath string) {
	b.requiredAssets.TexturePaths = append(b.requiredAssets.TexturePaths, texturePath)
}

// AddRequiredFont adds a font path to the required assets.
// Convenience method for adding fonts one at a time.
func (b *BaseScene) AddRequiredFont(fontPath string) {
	b.requiredAssets.FontPaths = append(b.requiredAssets.FontPaths, fontPath)
}

// ===== SceneStateful Interface =====

// SaveState saves the current scene state before cleanup.
// Default implementation returns the internal saved state map.
// Override this to save custom state.
// Implements SceneStateful interface.
func (b *BaseScene) SaveState() {
	// Default: no-op
	// Subclasses can override to populate b.savedState
}

// RestoreState restores the previously saved scene state after initialization.
// Default implementation is a no-op.
// Override this to restore custom state.
// Implements SceneStateful interface.
func (b *BaseScene) RestoreState() {
	// Default: no-op
	// Subclasses can override to read from b.savedState
}

// GetSavedState returns the internal saved state map.
// Use this in your custom SaveState/RestoreState implementations.
func (b *BaseScene) GetSavedState() map[string]interface{} {
	return b.savedState
}

// ===== SceneOverlayRenderer Interface =====

// RenderOverlays renders additional overlays (menus, HUD, debug) each frame.
// Default implementation renders the debug console if enabled.
// Override this to render custom overlays, and call b.BaseScene.RenderOverlays()
// at the end to include debug console.
// Implements SceneOverlayRenderer interface.
func (b *BaseScene) RenderOverlays() error {
	// Render debug console (engine feature)
	if err := b.RenderDebugConsole(); err != nil {
		return err
	}
	return nil
}

// ===== Debug Console (Engine Feature) =====

// initDebugConsole initializes the debug console for this scene.
// Called automatically by Initialize() if debug mode is enabled.
// Internal method - scenes should not call this directly.
func (b *BaseScene) initDebugConsole() error {
	if b.canvasManager == nil {
		return nil // Canvas not yet available, will retry later
	}

	logger.Logger.Debugf("Initializing debug console for %s scene", b.name)

	// Create and load font metadata
	b.debugFont = text.NewSpriteFont()
	err := b.debugFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load debug font: %s", err)
		return err
	}

	// Create text renderer
	b.debugTextRenderer = text.NewTextRenderer(b.canvasManager)
	b.debugConsoleInitialized = true

	logger.Logger.Debugf("Debug console initialized successfully for %s", b.name)
	debug.Console.PostMessage("System", b.name+" scene ready")

	return nil
}

// RenderDebugConsole renders the debug console overlay.
// This is called automatically by RenderOverlays() if debug mode is enabled.
// Scenes can override RenderOverlays() to customize rendering order.
func (b *BaseScene) RenderDebugConsole() error {
	if !config.Global.Debug.Enabled {
		return nil
	}

	// Lazy initialization if canvas wasn't available during Initialize()
	if !b.debugConsoleInitialized && b.canvasManager != nil {
		if err := b.initDebugConsole(); err != nil {
			return err
		}
	}

	if !b.debugConsoleInitialized || b.debugFont == nil || b.debugTextRenderer == nil {
		return nil // Not yet initialized
	}

	return debug.Console.Render(b.canvasManager, b.debugTextRenderer, b.debugFont)
}

// GetDebugFont returns the debug font (for scenes that need direct access).
func (b *BaseScene) GetDebugFont() text.Font {
	return b.debugFont
}

// ===== Layer Management Helper Methods =====

// AddBackground adds a GameObject to the BACKGROUND layer.
// This is a convenience method for layer management.
func (b *BaseScene) AddBackground(obj types.GameObject) {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()
	b.layers[BACKGROUND] = append(b.layers[BACKGROUND], obj)
}

// AddEntity adds a GameObject to the ENTITIES layer.
// This is a convenience method for layer management.
func (b *BaseScene) AddEntity(obj types.GameObject) {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()
	b.layers[ENTITIES] = append(b.layers[ENTITIES], obj)
}

// AddUI adds a GameObject to the UI layer.
// This is a convenience method for layer management.
func (b *BaseScene) AddUI(obj types.GameObject) {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()
	b.layers[UI] = append(b.layers[UI], obj)
}

// AddGameObject adds a GameObject to the specified layer.
// Use this for dynamic layer selection, or use the type-specific helpers above.
func (b *BaseScene) AddGameObject(layer SceneLayer, obj types.GameObject) {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()
	b.layers[layer] = append(b.layers[layer], obj)
}

// RemoveGameObject removes a GameObject by ID from all layers.
// Returns true if the object was found and removed.
func (b *BaseScene) RemoveGameObject(id string) bool {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()

	found := false
	for layer := range b.layers {
		filtered := []types.GameObject{}
		for _, obj := range b.layers[layer] {
			if obj.GetID() != id {
				filtered = append(filtered, obj)
			} else {
				found = true
			}
		}
		b.layers[layer] = filtered
	}

	return found
}

// GetGameObject finds a GameObject by ID across all layers.
// Returns nil if not found.
func (b *BaseScene) GetGameObject(id string) types.GameObject {
	b.layerMutex.RLock()
	defer b.layerMutex.RUnlock()

	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
		for _, obj := range b.layers[layer] {
			if obj.GetID() == id {
				return obj
			}
		}
	}

	return nil
}

// ClearLayer removes all GameObjects from the specified layer.
func (b *BaseScene) ClearLayer(layer SceneLayer) {
	b.layerMutex.Lock()
	defer b.layerMutex.Unlock()
	b.layers[layer] = []types.GameObject{}
}

// GetScreenWidth returns the scene's screen width.
func (b *BaseScene) GetScreenWidth() float64 {
	return b.screenWidth
}

// GetScreenHeight returns the scene's screen height.
func (b *BaseScene) GetScreenHeight() float64 {
	return b.screenHeight
}

// ===== SceneInjectable Interface =====

// InjectDependencies receives all engine dependencies in a single call.
// This is called automatically by the engine before Initialize().
// The deps parameter implements types.DependencyProvider interface.
// Implements SceneInjectable interface.
func (b *BaseScene) InjectDependencies(deps types.DependencyProvider) {
	// Use the DependencyProvider interface to access all dependencies
	// This avoids circular imports while maintaining type safety
	b.inputCapturer = deps.GetInputCapturer()
	b.stateChangeCallback = deps.GetStateChangeCallback()
	b.gameStateManager = deps.GetGameStateProvider()
	b.screenWidth = deps.GetScreenWidth()
	b.screenHeight = deps.GetScreenHeight()

	// Canvas manager needs type assertion from interface{}
	if cm, ok := deps.GetCanvasManager().(canvas.CanvasManager); ok {
		b.canvasManager = cm
	}
}

// GetLayer returns the game objects in the specified layer (thread-safe read).
// Returns a copy to prevent external modification.
func (b *BaseScene) GetLayer(layer SceneLayer) []types.GameObject {
	b.layerMutex.RLock()
	defer b.layerMutex.RUnlock()
	// Return a copy to prevent external modification
	result := make([]types.GameObject, len(b.layers[layer]))
	copy(result, b.layers[layer])
	return result
}

// GetInputCapturer returns the input capturer (for backwards compatibility with scenes that need direct access).
func (b *BaseScene) GetInputCapturer() types.InputCapturer {
	return b.inputCapturer
}

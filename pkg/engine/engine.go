//go:build js

package engine

import (
	"fmt"
	"sync"
	"syscall/js"

	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/input"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Engine represents the game engine that manages the canvas and game loop
type Engine struct {
	canvasManager      canvas.CanvasManager
	inputCapturer      types.InputCapturer
	lastTime           float64
	textureLoaded      bool
	running            bool
	currentGameState   types.GameState
	gameStatePipelines map[types.GameState][]types.PipelineType
	registeredScenes   map[types.GameState]scene.Scene // Scenes registered by external users
	currentScene       scene.Scene                     // Active scene managing game objects
	stateLock          sync.Mutex                      // Lock to prevent concurrent state changes
	screenWidth        float64
	screenHeight       float64
}

// NewEngine creates a new game engine instance
func NewEngine() *Engine {
	e := &Engine{
		canvasManager:      canvas.NewWebGPUCanvasManager(),
		inputCapturer:      input.NewUnifiedInput(),
		running:            false,
		gameStatePipelines: make(map[types.GameState][]types.PipelineType),
		registeredScenes:   make(map[types.GameState]scene.Scene),
		screenWidth:        config.Global.Screen.Width,
		screenHeight:       config.Global.Screen.Height,
	}

	// Initialize game state pipeline mappings
	e.initializeGameStates()

	return e
}

// initializeGameStates sets up the pipeline configurations for each game state
func (e *Engine) initializeGameStates() {
	// GAMEPLAY state uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.GAMEPLAY] = []types.PipelineType{
		types.TexturedPipeline,
	}
	// BATTLE state also uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.BATTLE] = []types.PipelineType{
		types.TexturedPipeline,
	}
}

// RegisterScene registers a scene for a specific game state
func (e *Engine) RegisterScene(state types.GameState, scene scene.Scene) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	e.registeredScenes[state] = scene
	logger.Logger.Debugf("Registered scene for game state: %s", state.String())
}

// Initialize sets up the engine with the specified canvas ID
func (e *Engine) Initialize(canvasID string) error {
	logger.Logger.Debugf("Engine initializing with canvas: %s", canvasID)

	// Register debug console as global debug poster
	types.SetGlobalDebugPoster(debug.Console)

	err := e.canvasManager.Initialize(canvasID)
	if err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return err
	}

	// Initialize input capturer
	err = e.inputCapturer.Initialize()
	if err != nil {
		logger.Logger.Errorf("Failed to initialize input: %s", err.Error())
		return err
	}

	logger.Logger.Debugf("Engine initialized successfully")
	return nil
}

// Start begins th e game loop
func (e *Engine) Start() {
	if e.running {
		logger.Logger.Debugf("Engine already running")
		return
	}

	e.running = true
	logger.Logger.Debugf("Engine starting render loop")

	e.startRenderLoop()
}

// startRenderLoop initializes and starts the animation loop
func (e *Engine) startRenderLoop() {
	var animationLoop js.Func
	animationLoop = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !e.running {
			return nil
		}

		currentTime := js.Global().Get("performance").Call("now").Float() / 1000.0

		if e.lastTime == 0 {
			e.lastTime = currentTime
		}

		deltaTime := currentTime - e.lastTime
		e.lastTime = currentTime

		// Update and render the frame
		e.Update(deltaTime)
		e.Render()

		js.Global().Call("requestAnimationFrame", animationLoop)
		return nil
	})

	js.Global().Call("requestAnimationFrame", animationLoop)
}

// Update handles game logic updates
func (e *Engine) Update(deltaTime float64) {
	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	// Delegate update to the current scene
	if currentScene != nil {
		currentScene.Update(deltaTime)
	}

	// TODO: This should be done in a separate thread,
	// at game state change, with a black screen in between.
	// Load textures for sprites if needed
	e.loadSpriteTextures()
}

// Render draws the current frame
func (e *Engine) Render() {
	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	// Get renderables from scene in correct layer order
	var renderables []types.GameObject
	if currentScene != nil {
		renderables = currentScene.GetRenderables()
	}

	// Check if we have anything to render
	if len(renderables) > 0 {
		err := e.canvasManager.BeginBatch()
		if err != nil {
			logger.Logger.Errorf("Failed to begin batch: %s", err.Error())
		}

		// Render all game objects in layer order
		for _, gameObject := range renderables {
			var renderData types.SpriteRenderData
			if mover := gameObject.GetMover(); mover != nil {
				renderData = gameObject.GetSprite().GetSpriteRenderData(mover.GetPosition())
			} else {
				renderData = gameObject.GetSprite().GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
			}

			if !renderData.Visible {
				continue
			}

			err := e.canvasManager.DrawTexturedRect(
				renderData.TexturePath,
				renderData.Position,
				renderData.Size,
				renderData.UV,
			)
			if err != nil {
				// Texture might not be loaded yet
				continue
			}
		}

		// Render scene-specific overlays (if implemented) inside batch
		if overlayRenderer, ok := currentScene.(types.SceneOverlayRenderer); ok {
			if err := overlayRenderer.RenderOverlays(); err != nil {
				logger.Logger.Tracef("Failed to render scene overlays: %s", err.Error())
			}
		}

		err = e.canvasManager.EndBatch()
		if err != nil {
			logger.Logger.Errorf("Failed to end batch: %s", err.Error())
		}
	}

	e.canvasManager.Render()
}

// loadSpriteTextures loads textures for all game objects in the current scene
func (e *Engine) loadSpriteTextures() {
	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	if currentScene == nil {
		return
	}

	// Get all renderables from the scene
	renderables := currentScene.GetRenderables()

	// Load textures for all game objects
	for _, gameObject := range renderables {
		pos := types.Vector2{X: 0, Y: 0}
		if mover := gameObject.GetMover(); mover != nil {
			pos = mover.GetPosition()
		}
		renderData := gameObject.GetSprite().GetSpriteRenderData(pos)
		e.canvasManager.LoadTexture(renderData.TexturePath)
	}

	// Load any extra textures requested by the scene
	if textureProvider, ok := currentScene.(types.SceneTextureProvider); ok {
		for _, path := range textureProvider.GetExtraTexturePaths() {
			if path != "" {
				e.canvasManager.LoadTexture(path)
			}
		}
	}
}

// Stop stops the game loop
func (e *Engine) Stop() {
	e.running = false
	logger.Logger.Debugf("Engine stopped")
}

// Cleanup releases engine resources
func (e *Engine) Cleanup() error {
	e.Stop()

	// Cleanup input capturer
	if e.inputCapturer != nil {
		e.inputCapturer.Cleanup()
	}

	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager for advanced usage
func (e *Engine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}

// SetGameState changes the current game state and updates the active pipelines
func (e *Engine) SetGameState(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	pipelines, exists := e.gameStatePipelines[state]
	if !exists {
		return &EngineError{Message: "Game state not configured: " + state.String()}
	}

	err := e.canvasManager.SetPipelines(pipelines)
	if err != nil {
		return err
	}

	// Save state of old scene before cleanup if it's stateful
	if e.currentScene != nil {
		if stateful, ok := e.currentScene.(types.SceneStateful); ok {
			stateful.SaveState()
			logger.Logger.Debugf("Saved state for scene: %s", e.currentScene.GetName())
		}
		e.currentScene.Cleanup()
		e.currentScene = nil
	}

	// Get registered scene for this state
	registeredScene, exists := e.registeredScenes[state]
	if !exists {
		return &EngineError{Message: "No scene registered for game state: " + state.String()}
	}

	// Preload all scene assets BEFORE initialization to avoid deadlocks
	// This ensures all blocking I/O (fonts, textures) happens before the game loop
	err = e.preloadSceneAssets(registeredScene)
	if err != nil {
		logger.Logger.Warnf("Some assets failed to preload for scene %s: %s", registeredScene.GetName(), err.Error())
		// Continue anyway - assets might load lazily later
	}

	// Inject input capturer if scene implements SceneInputProvider
	if inputProvider, ok := registeredScene.(types.SceneInputProvider); ok {
		inputProvider.SetInputCapturer(e.inputCapturer)
		logger.Logger.Debugf("Injected input capturer into scene: %s", registeredScene.GetName())
	}

	// Inject state change callback if scene implements SceneStateChangeRequester
	if stateRequester, ok := registeredScene.(types.SceneStateChangeRequester); ok {
		stateRequester.SetStateChangeCallback(e.SetGameState)
		logger.Logger.Debugf("Injected state change callback into scene: %s", registeredScene.GetName())
	}

	// Initialize the registered scene (assets are already loaded, so this should be fast)
	err = registeredScene.Initialize()
	if err != nil {
		return &EngineError{Message: "Failed to initialize scene: " + err.Error()}
	}

	// Restore state if scene is stateful (after initialization so objects exist)
	if stateful, ok := registeredScene.(types.SceneStateful); ok {
		stateful.RestoreState()
		logger.Logger.Debugf("Restored state for scene: %s", registeredScene.GetName())
	}

	e.currentScene = registeredScene
	logger.Logger.Debugf("Initialized scene: %s for game state: %s", registeredScene.GetName(), state.String())

	e.currentGameState = state
	logger.Logger.Debugf("Game state changed to: %s", state.String())
	return nil
}

// preloadSceneAssets loads all assets required by a scene before initialization
// This prevents deadlocks by doing all blocking I/O upfront, before the game loop
func (e *Engine) preloadSceneAssets(s scene.Scene) error {
	logger.Logger.Debugf("Preloading assets for scene: %s", s.GetName())

	var errors []error

	// Check if scene implements SceneAssetProvider to get asset list
	if assetProvider, ok := s.(types.SceneAssetProvider); ok {
		assets := assetProvider.GetRequiredAssets()

		// Preload all textures
		for _, texturePath := range assets.TexturePaths {
			if texturePath != "" {
				e.canvasManager.LoadTexture(texturePath)
				logger.Logger.Debugf("Preloaded texture: %s", texturePath)
			}
		}

		// Preload all fonts (blocking operations - this is where deadlocks used to occur)
		// Font cache will prevent duplicate loads across scenes
		for _, fontPath := range assets.FontPaths {
			if fontPath != "" {
				// Create a temporary font and load it to ensure it's ready
				// The scene will reuse this font instance via cache
				tempFont := text.NewSpriteFont()
				err := tempFont.LoadFont(fontPath)
				if err != nil {
					errMsg := fmt.Errorf("failed to preload font %s: %w", fontPath, err)
					logger.Logger.Warnf("%s", errMsg.Error())
					errors = append(errors, errMsg)
					// Continue loading other assets
				} else {
					logger.Logger.Debugf("Preloaded font: %s (cached for reuse)", fontPath)
				}
			}
		}
	} else {
		// Fallback: try to discover assets from SceneTextureProvider
		if textureProvider, ok := s.(types.SceneTextureProvider); ok {
			for _, path := range textureProvider.GetExtraTexturePaths() {
				if path != "" {
					e.canvasManager.LoadTexture(path)
					logger.Logger.Debugf("Preloaded texture (fallback): %s", path)
				}
			}
		}
	}

	// Log summary
	if len(errors) > 0 {
		logger.Logger.Warnf("Preloaded assets for scene %s with %d error(s)", s.GetName(), len(errors))
		return fmt.Errorf("preload completed with %d error(s)", len(errors))
	}

	logger.Logger.Debugf("Successfully preloaded all assets for scene: %s", s.GetName())
	return nil
}

// GetGameState returns the current game state
func (e *Engine) GetGameState() types.GameState {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	return e.currentGameState
}

// EngineError represents an error in the engine
type EngineError struct {
	Message string
}

func (e *EngineError) Error() string {
	return e.Message
}

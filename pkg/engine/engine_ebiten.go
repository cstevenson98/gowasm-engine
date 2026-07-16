package engine

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/input"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// EbitenEngine represents the game engine that manages the canvas and game loop for Ebiten
// It implements the ebiten.Game interface
type EbitenEngine struct {
	canvasManager      *canvas.EbitenCanvasManager
	inputCapturer      *input.EbitenInput
	running            bool
	currentGameState   types.GameState
	gameStatePipelines map[types.GameState][]types.PipelineType
	registeredScenes   map[types.GameState]scene.Scene
	currentScene       scene.Scene
	stateLock          sync.Mutex
	screenWidth        float64
	screenHeight       float64
	gameStateProvider  interface{}
	
	// Ebiten-specific
	tickCount uint64 // Frame counter
}

// NewEbitenEngine creates a new Ebiten game engine instance
func NewEbitenEngine() *EbitenEngine {
	e := &EbitenEngine{
		canvasManager:      canvas.NewEbitenCanvasManager(),
		inputCapturer:      input.NewEbitenInput(),
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
func (e *EbitenEngine) initializeGameStates() {
	// MENU state uses textured pipeline for text rendering
	e.gameStatePipelines[types.MENU] = []types.PipelineType{
		types.TexturedPipeline,
	}
	// GAMEPLAY state uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.GAMEPLAY] = []types.PipelineType{
		types.TexturedPipeline,
	}
	// PLAYER_MENU state uses textured pipeline for text rendering
	e.gameStatePipelines[types.PLAYER_MENU] = []types.PipelineType{
		types.TexturedPipeline,
	}
	// BATTLE state also uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.BATTLE] = []types.PipelineType{
		types.TexturedPipeline,
	}
}

// RegisterScene registers a scene for a specific game state
func (e *EbitenEngine) RegisterScene(state types.GameState, scene scene.Scene) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	e.registeredScenes[state] = scene
	logger.Logger.Debugf("Registered scene for game state: %s", state.String())
}

// Initialize sets up the engine
func (e *EbitenEngine) Initialize(canvasID string) error {
	logger.Logger.Debugf("EbitenEngine initializing")

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

	logger.Logger.Debugf("EbitenEngine initialized successfully")
	return nil
}

// Start begins the game loop (called before ebiten.RunGame)
func (e *EbitenEngine) Start() {
	if e.running {
		logger.Logger.Debugf("Engine already running")
		return
	}

	e.running = true
	logger.Logger.Debugf("EbitenEngine started")
}

// Update implements ebiten.Game interface - called 60 times per second
func (e *EbitenEngine) Update() error {
	if !e.running {
		return nil
	}

	e.tickCount++
	
	// Fixed timestep: Ebiten runs at 60 TPS by default
	deltaTime := 1.0 / 60.0

	// Poll input
	e.inputCapturer.PollInput()

	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	// Delegate update to the current scene
	if currentScene != nil {
		currentScene.Update(deltaTime)
	}

	// Load textures for sprites if needed
	e.loadSpriteTextures()

	// Check for ESC key to quit (optional debug feature)
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

// Draw implements ebiten.Game interface - renders the current frame
func (e *EbitenEngine) Draw(screen *ebiten.Image) {
	// Set the screen in canvas manager
	e.canvasManager.SetScreen(screen)

	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	// Get renderables from scene in correct layer order
	var renderables []types.GameObject
	if currentScene != nil {
		renderables = currentScene.GetRenderables()
	}

	// Check if we have anything to render OR if scene has overlays to render
	hasOverlays := false
	if currentScene != nil {
		_, hasOverlays = currentScene.(types.SceneOverlayRenderer)
	}

	if len(renderables) > 0 || hasOverlays {
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
}

// Layout implements ebiten.Game interface - returns the virtual screen size
func (e *EbitenEngine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Return virtual resolution - Ebiten will scale automatically
	return int(e.screenWidth), int(e.screenHeight)
}

// loadSpriteTextures loads textures for all game objects in the current scene
func (e *EbitenEngine) loadSpriteTextures() {
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
func (e *EbitenEngine) Stop() {
	e.running = false
	logger.Logger.Debugf("EbitenEngine stopped")
}

// Cleanup releases engine resources
func (e *EbitenEngine) Cleanup() error {
	e.Stop()

	// Cleanup input capturer
	if e.inputCapturer != nil {
		e.inputCapturer.Cleanup()
	}

	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager for advanced usage
func (e *EbitenEngine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}

// RegisterGameStateProvider registers a game state provider that will be injected
// into scenes that implement SceneGameStateUser
func (e *EbitenEngine) RegisterGameStateProvider(provider interface{}) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	e.gameStateProvider = provider
	logger.Logger.Debugf("Registered game state provider with engine")
}

// SetGameState changes the current game state and updates the active pipelines
func (e *EbitenEngine) SetGameState(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	pipelines, exists := e.gameStatePipelines[state]
	if !exists {
		return fmt.Errorf("game state not configured: %s", state.String())
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
		return fmt.Errorf("no scene registered for game state: %s", state.String())
	}

	// Preload all scene assets BEFORE initialization
	err = e.preloadSceneAssets(registeredScene)
	if err != nil {
		logger.Logger.Warnf("Some assets failed to preload for scene %s: %s", registeredScene.GetName(), err.Error())
	}

	// Inject dependencies
	if injectable, ok := registeredScene.(types.SceneInjectable); ok {
		injectable.InjectDependencies(e.GetDependencies())
		logger.Logger.Debugf("Injected all dependencies into scene via InjectDependencies(): %s", registeredScene.GetName())
	} else {
		// Fallback to manual injection
		logger.Logger.Debugf("Using manual dependency injection for scene: %s", registeredScene.GetName())
		
		if inputProvider, ok := registeredScene.(types.SceneInputProvider); ok {
			inputProvider.SetInputCapturer(e.inputCapturer)
		}

		if stateRequester, ok := registeredScene.(types.SceneStateChangeRequester); ok {
			stateRequester.SetStateChangeCallback(e.SetGameState)
		}

		if gameStateUser, ok := registeredScene.(types.SceneGameStateUser); ok {
			if e.gameStateProvider != nil {
				gameStateUser.SetGameState(e.gameStateProvider)
			}
		}
		
		// Inject canvas manager
		type canvasManagerSetter interface {
			SetCanvasManager(canvas.CanvasManager)
		}
		if cmSetter, ok := registeredScene.(canvasManagerSetter); ok {
			cmSetter.SetCanvasManager(e.canvasManager)
		}
	}

	// Initialize the registered scene
	err = registeredScene.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize scene: %w", err)
	}

	// Restore state if scene is stateful
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
func (e *EbitenEngine) preloadSceneAssets(s scene.Scene) error {
	logger.Logger.Debugf("Preloading assets for scene: %s", s.GetName())

	var errors []error

	// Check if scene implements SceneAssetProvider
	if assetProvider, ok := s.(types.SceneAssetProvider); ok {
		assets := assetProvider.GetRequiredAssets()

		// Preload all textures
		for _, texturePath := range assets.TexturePaths {
			if texturePath != "" {
				err := e.canvasManager.LoadTexture(texturePath)
				if err != nil {
					errors = append(errors, err)
				} else {
					logger.Logger.Debugf("Preloaded texture: %s", texturePath)
				}
			}
		}

		// Preload all fonts
		for _, fontPath := range assets.FontPaths {
			if fontPath != "" {
				tempFont := text.NewSpriteFont()
				err := tempFont.LoadFont(fontPath)
				if err != nil {
					errMsg := fmt.Errorf("failed to preload font %s: %w", fontPath, err)
					logger.Logger.Warnf("%s", errMsg.Error())
					errors = append(errors, errMsg)
				} else {
					logger.Logger.Debugf("Preloaded font: %s (cached for reuse)", fontPath)
					
					// Also load the font's texture
					fontTexturePath := tempFont.GetTexturePath()
					if fontTexturePath != "" {
						err := e.canvasManager.LoadTexture(fontTexturePath)
						if err != nil {
							errMsg := fmt.Errorf("failed to preload font texture %s: %w", fontTexturePath, err)
							logger.Logger.Warnf("%s", errMsg.Error())
							errors = append(errors, errMsg)
						} else {
							logger.Logger.Debugf("Preloaded font texture: %s", fontTexturePath)
						}
					}
				}
			}
		}
	} else {
		// Fallback: try SceneTextureProvider
		if textureProvider, ok := s.(types.SceneTextureProvider); ok {
			for _, path := range textureProvider.GetExtraTexturePaths() {
				if path != "" {
					e.canvasManager.LoadTexture(path)
					logger.Logger.Debugf("Preloaded texture (fallback): %s", path)
				}
			}
		}
	}

	if len(errors) > 0 {
		logger.Logger.Warnf("Preloaded assets for scene %s with %d error(s)", s.GetName(), len(errors))
		return fmt.Errorf("preload completed with %d error(s)", len(errors))
	}

	logger.Logger.Debugf("Successfully preloaded all assets for scene: %s", s.GetName())
	return nil
}

// GetGameState returns the current game state
func (e *EbitenEngine) GetGameState() types.GameState {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	return e.currentGameState
}

// EngineDependencies holds all injectable dependencies from the engine (Ebiten version)
type EngineDependencies struct {
	InputCapturer       *input.EbitenInput
	CanvasManager       *canvas.EbitenCanvasManager
	StateChangeCallback func(state types.GameState) error
	GameStateProvider   interface{}
	ScreenWidth         float64
	ScreenHeight        float64
}

// GetDependencies returns all injectable dependencies for scenes
func (e *EbitenEngine) GetDependencies() *EngineDependencies {
	return &EngineDependencies{
		InputCapturer:       e.inputCapturer,
		CanvasManager:       e.canvasManager,
		StateChangeCallback: e.SetGameState,
		GameStateProvider:   e.gameStateProvider,
		ScreenWidth:         e.screenWidth,
		ScreenHeight:        e.screenHeight,
	}
}

// Implement types.DependencyProvider interface methods

// GetInputCapturer returns the input capturer
func (d *EngineDependencies) GetInputCapturer() types.InputCapturer {
	return d.InputCapturer
}

// GetCanvasManager returns the canvas manager as interface{}
func (d *EngineDependencies) GetCanvasManager() interface{} {
	return d.CanvasManager
}

// GetStateChangeCallback returns the state change callback
func (d *EngineDependencies) GetStateChangeCallback() func(types.GameState) error {
	return d.StateChangeCallback
}

// GetGameStateProvider returns the game state provider
func (d *EngineDependencies) GetGameStateProvider() interface{} {
	return d.GameStateProvider
}

// GetScreenWidth returns the screen width
func (d *EngineDependencies) GetScreenWidth() float64 {
	return d.ScreenWidth
}

// GetScreenHeight returns the screen height
func (d *EngineDependencies) GetScreenHeight() float64 {
	return d.ScreenHeight
}

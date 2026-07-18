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

// Engine manages the canvas, input, and game loop. It implements the
// ebiten.Game interface (Update, Draw, Layout).
type Engine struct {
	canvasManager *canvas.Canvas
	inputCapturer *input.Input

	running           bool
	currentGameState  types.GameState
	registeredScenes  map[types.GameState]scene.Scene
	currentScene      scene.Scene
	stateLock         sync.Mutex
	screenWidth       float64
	screenHeight      float64
	gameStateProvider interface{}

	tickCount uint64
}

// NewEngine creates a new game engine instance.
func NewEngine() *Engine {
	return &Engine{
		canvasManager:    canvas.NewCanvas(),
		inputCapturer:    input.NewInput(),
		running:          false,
		registeredScenes: make(map[types.GameState]scene.Scene),
		screenWidth:      config.Global.Screen.Width,
		screenHeight:     config.Global.Screen.Height,
	}
}

// RegisterScene registers a scene for a specific game state.
func (e *Engine) RegisterScene(state types.GameState, s scene.Scene) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	e.registeredScenes[state] = s
	logger.Logger.Debugf("Registered scene for game state: %s", state.String())
}

// Initialize sets up the engine.
func (e *Engine) Initialize(canvasID string) error {
	logger.Logger.Debugf("Engine initializing")

	// Register debug console as global debug poster
	types.SetGlobalDebugPoster(debug.Console)

	if err := e.canvasManager.Initialize(canvasID); err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return err
	}

	if err := e.inputCapturer.Initialize(); err != nil {
		logger.Logger.Errorf("Failed to initialize input: %s", err.Error())
		return err
	}

	logger.Logger.Debugf("Engine initialized successfully")
	return nil
}

// Start begins the game loop (called before ebiten.RunGame).
func (e *Engine) Start() {
	if e.running {
		logger.Logger.Debugf("Engine already running")
		return
	}

	e.running = true
	logger.Logger.Debugf("Engine started")
}

// Update implements ebiten.Game - called 60 times per second.
func (e *Engine) Update() error {
	if !e.running {
		return nil
	}

	e.tickCount++

	// Fixed timestep: Ebiten runs at 60 TPS by default
	deltaTime := 1.0 / 60.0

	e.inputCapturer.PollInput()

	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	if currentScene != nil {
		currentScene.Update(deltaTime)
	}

	e.loadSpriteTextures()

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	return nil
}

// Draw implements ebiten.Game - renders the current frame.
func (e *Engine) Draw(screen *ebiten.Image) {
	e.canvasManager.SetScreen(screen)

	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	if currentScene == nil {
		return
	}

	// Render all game objects in layer order.
	for _, gameObject := range currentScene.GetRenderables() {
		var renderData types.SpriteRenderData
		if mover := gameObject.GetMover(); mover != nil {
			renderData = gameObject.GetSprite().GetSpriteRenderData(mover.GetPosition())
		} else {
			renderData = gameObject.GetSprite().GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
		}

		if !renderData.Visible {
			continue
		}

		// Texture may not be loaded yet; skip on error.
		_ = e.canvasManager.DrawTexturedRect(
			renderData.TexturePath,
			renderData.Position,
			renderData.Size,
			renderData.UV,
		)
	}

	// Render scene-specific overlays (menus, HUD, debug console).
	if overlayRenderer, ok := currentScene.(types.SceneOverlayRenderer); ok {
		if err := overlayRenderer.RenderOverlays(); err != nil {
			logger.Logger.Tracef("Failed to render scene overlays: %s", err.Error())
		}
	}
}

// Layout implements ebiten.Game - returns the virtual screen size.
func (e *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(e.screenWidth), int(e.screenHeight)
}

// loadSpriteTextures loads textures for all game objects in the current scene.
func (e *Engine) loadSpriteTextures() {
	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	if currentScene == nil {
		return
	}

	for _, gameObject := range currentScene.GetRenderables() {
		pos := types.Vector2{X: 0, Y: 0}
		if mover := gameObject.GetMover(); mover != nil {
			pos = mover.GetPosition()
		}
		renderData := gameObject.GetSprite().GetSpriteRenderData(pos)
		_ = e.canvasManager.LoadTexture(renderData.TexturePath)
	}

	if textureProvider, ok := currentScene.(types.SceneTextureProvider); ok {
		for _, path := range textureProvider.GetExtraTexturePaths() {
			if path != "" {
				_ = e.canvasManager.LoadTexture(path)
			}
		}
	}
}

// Stop stops the game loop.
func (e *Engine) Stop() {
	e.running = false
	logger.Logger.Debugf("Engine stopped")
}

// Cleanup releases engine resources.
func (e *Engine) Cleanup() error {
	e.Stop()

	if e.inputCapturer != nil {
		e.inputCapturer.Cleanup()
	}

	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager.
func (e *Engine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}

// RegisterGameStateProvider registers a game state provider that will be
// injected into scenes that implement SceneGameStateUser.
func (e *Engine) RegisterGameStateProvider(provider interface{}) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	e.gameStateProvider = provider
	logger.Logger.Debugf("Registered game state provider with engine")
}

// SetGameState changes the current game state and switches the active scene.
func (e *Engine) SetGameState(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	registeredScene, exists := e.registeredScenes[state]
	if !exists {
		return fmt.Errorf("no scene registered for game state: %s", state.String())
	}

	// Save state of old scene before cleanup if it's stateful.
	if e.currentScene != nil {
		if stateful, ok := e.currentScene.(types.SceneStateful); ok {
			stateful.SaveState()
			logger.Logger.Debugf("Saved state for scene: %s", e.currentScene.GetName())
		}
		e.currentScene.Cleanup()
		e.currentScene = nil
	}

	// Preload all scene assets BEFORE initialization.
	if err := e.preloadSceneAssets(registeredScene); err != nil {
		logger.Logger.Warnf("Some assets failed to preload for scene %s: %s", registeredScene.GetName(), err.Error())
	}

	// Inject dependencies.
	if injectable, ok := registeredScene.(types.SceneInjectable); ok {
		injectable.InjectDependencies(e.GetDependencies())
		logger.Logger.Debugf("Injected dependencies into scene: %s", registeredScene.GetName())
	} else {
		e.manualInject(registeredScene)
	}

	// Initialize the registered scene.
	if err := registeredScene.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize scene: %w", err)
	}

	// Restore state if scene is stateful.
	if stateful, ok := registeredScene.(types.SceneStateful); ok {
		stateful.RestoreState()
		logger.Logger.Debugf("Restored state for scene: %s", registeredScene.GetName())
	}

	e.currentScene = registeredScene
	e.currentGameState = state
	logger.Logger.Debugf("Game state changed to: %s", state.String())
	return nil
}

// manualInject provides dependency injection for scenes that don't embed BaseScene.
func (e *Engine) manualInject(s scene.Scene) {
	logger.Logger.Debugf("Using manual dependency injection for scene: %s", s.GetName())

	if inputProvider, ok := s.(types.SceneInputProvider); ok {
		inputProvider.SetInputCapturer(e.inputCapturer)
	}
	if stateRequester, ok := s.(types.SceneStateChangeRequester); ok {
		stateRequester.SetStateChangeCallback(e.SetGameState)
	}
	if gameStateUser, ok := s.(types.SceneGameStateUser); ok {
		if e.gameStateProvider != nil {
			gameStateUser.SetGameState(e.gameStateProvider)
		}
	}
	type canvasManagerSetter interface {
		SetCanvasManager(canvas.CanvasManager)
	}
	if cmSetter, ok := s.(canvasManagerSetter); ok {
		cmSetter.SetCanvasManager(e.canvasManager)
	}
}

// preloadSceneAssets loads all assets required by a scene before initialization.
func (e *Engine) preloadSceneAssets(s scene.Scene) error {
	logger.Logger.Debugf("Preloading assets for scene: %s", s.GetName())

	var errs []error

	if assetProvider, ok := s.(types.SceneAssetProvider); ok {
		assets := assetProvider.GetRequiredAssets()

		for _, texturePath := range assets.TexturePaths {
			if texturePath != "" {
				if err := e.canvasManager.LoadTexture(texturePath); err != nil {
					errs = append(errs, err)
				} else {
					logger.Logger.Debugf("Preloaded texture: %s", texturePath)
				}
			}
		}

		for _, fontPath := range assets.FontPaths {
			if fontPath == "" {
				continue
			}
			tempFont := text.NewSpriteFont()
			if err := tempFont.LoadFont(fontPath); err != nil {
				errMsg := fmt.Errorf("failed to preload font %s: %w", fontPath, err)
				logger.Logger.Warnf("%s", errMsg.Error())
				errs = append(errs, errMsg)
				continue
			}
			logger.Logger.Debugf("Preloaded font: %s (cached for reuse)", fontPath)

			if fontTexturePath := tempFont.GetTexturePath(); fontTexturePath != "" {
				if err := e.canvasManager.LoadTexture(fontTexturePath); err != nil {
					errMsg := fmt.Errorf("failed to preload font texture %s: %w", fontTexturePath, err)
					logger.Logger.Warnf("%s", errMsg.Error())
					errs = append(errs, errMsg)
				} else {
					logger.Logger.Debugf("Preloaded font texture: %s", fontTexturePath)
				}
			}
		}
	} else if textureProvider, ok := s.(types.SceneTextureProvider); ok {
		for _, path := range textureProvider.GetExtraTexturePaths() {
			if path != "" {
				_ = e.canvasManager.LoadTexture(path)
				logger.Logger.Debugf("Preloaded texture (fallback): %s", path)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("preload completed with %d error(s)", len(errs))
	}

	logger.Logger.Debugf("Successfully preloaded all assets for scene: %s", s.GetName())
	return nil
}

// GetGameState returns the current game state.
func (e *Engine) GetGameState() types.GameState {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	return e.currentGameState
}

// EngineDependencies holds all injectable dependencies from the engine.
type EngineDependencies struct {
	InputCapturer       *input.Input
	CanvasManager       *canvas.Canvas
	StateChangeCallback func(state types.GameState) error
	GameStateProvider   interface{}
	ScreenWidth         float64
	ScreenHeight        float64
}

// GetDependencies returns all injectable dependencies for scenes.
func (e *Engine) GetDependencies() *EngineDependencies {
	return &EngineDependencies{
		InputCapturer:       e.inputCapturer,
		CanvasManager:       e.canvasManager,
		StateChangeCallback: e.SetGameState,
		GameStateProvider:   e.gameStateProvider,
		ScreenWidth:         e.screenWidth,
		ScreenHeight:        e.screenHeight,
	}
}

// Implement types.DependencyProvider interface.

// GetInputCapturer returns the input capturer.
func (d *EngineDependencies) GetInputCapturer() types.InputCapturer {
	return d.InputCapturer
}

// GetCanvasManager returns the canvas manager as interface{}.
func (d *EngineDependencies) GetCanvasManager() interface{} {
	return d.CanvasManager
}

// GetStateChangeCallback returns the state change callback.
func (d *EngineDependencies) GetStateChangeCallback() func(types.GameState) error {
	return d.StateChangeCallback
}

// GetGameStateProvider returns the game state provider.
func (d *EngineDependencies) GetGameStateProvider() interface{} {
	return d.GameStateProvider
}

// GetScreenWidth returns the screen width.
func (d *EngineDependencies) GetScreenWidth() float64 {
	return d.ScreenWidth
}

// GetScreenHeight returns the screen height.
func (d *EngineDependencies) GetScreenHeight() float64 {
	return d.ScreenHeight
}

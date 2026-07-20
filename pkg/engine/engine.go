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

	// Pending scene switch. Requests made during a scene's Update are recorded
	// here and applied at the start of the next frame, so a scene is never torn
	// down while it is still running its own Update (see RequestStateChange).
	pendingState    types.GameState
	hasPendingState bool
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

	// Fixed timestep: Ebiten runs at 60 TPS by default
	deltaTime := 1.0 / 60.0

	e.inputCapturer.PollInput()

	// Apply any scene switch requested during the previous frame before running
	// the current scene's Update, so switches never happen mid-Update.
	e.applyPendingStateChange()

	e.stateLock.Lock()
	currentScene := e.currentScene
	e.stateLock.Unlock()

	if currentScene != nil {
		currentScene.Update(deltaTime)
	}

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
		renderData := gameObject.GetRenderData()
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

// SetGameState changes the current game state and switches the active scene
// immediately. Use this to activate the starting scene before the game loop
// runs. Scenes should not call this from within Update - use the injected
// state-change callback (RequestStateChange) instead, which defers the switch.
func (e *Engine) SetGameState(state types.GameState) error {
	return e.applyGameState(state)
}

// RequestStateChange asks the engine to switch to state on the next frame. It
// is the callback injected into scenes, so a scene can request a switch from
// within its own Update without being torn down mid-Update. The switch is
// applied by applyPendingStateChange at the start of the next frame.
func (e *Engine) RequestStateChange(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	if _, exists := e.registeredScenes[state]; !exists {
		return fmt.Errorf("no scene registered for game state: %s", state.String())
	}

	e.pendingState = state
	e.hasPendingState = true
	logger.Logger.Debugf("Requested game state change to: %s", state.String())
	return nil
}

// applyPendingStateChange performs a deferred scene switch, if one was
// requested. Called by Update at the start of each frame.
func (e *Engine) applyPendingStateChange() {
	e.stateLock.Lock()
	if !e.hasPendingState {
		e.stateLock.Unlock()
		return
	}
	state := e.pendingState
	e.hasPendingState = false
	e.stateLock.Unlock()

	if err := e.applyGameState(state); err != nil {
		logger.Logger.Errorf("Failed to apply pending state change to %s: %s", state.String(), err.Error())
	}
}

// applyGameState performs the actual scene switch: tear down the current scene,
// preload the new scene's assets, inject dependencies, and initialize it.
func (e *Engine) applyGameState(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	registeredScene, exists := e.registeredScenes[state]
	if !exists {
		return fmt.Errorf("no scene registered for game state: %s", state.String())
	}

	// Tear down the outgoing scene.
	if e.currentScene != nil {
		e.currentScene.Cleanup()
		e.currentScene = nil
	}

	// Preload all scene assets BEFORE initialization.
	if err := e.preloadSceneAssets(registeredScene); err != nil {
		logger.Logger.Warnf("Some assets failed to preload for scene %s: %s", registeredScene.GetName(), err.Error())
	}

	// Inject dependencies. Scenes embed BaseScene (or otherwise implement
	// SceneInjectable) to receive engine services in a single call.
	if injectable, ok := registeredScene.(types.SceneInjectable); ok {
		injectable.InjectDependencies(e.GetDependencies())
		logger.Logger.Debugf("Injected dependencies into scene: %s", registeredScene.GetName())
	} else {
		logger.Logger.Warnf("Scene %s does not implement SceneInjectable; no dependencies injected", registeredScene.GetName())
	}

	// Initialize the registered scene.
	if err := registeredScene.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize scene: %w", err)
	}

	e.currentScene = registeredScene
	e.currentGameState = state
	logger.Logger.Debugf("Game state changed to: %s", state.String())
	return nil
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
		StateChangeCallback: e.RequestStateChange,
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

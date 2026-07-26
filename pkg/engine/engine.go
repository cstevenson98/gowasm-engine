package engine

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/milo/pkg/canvas"
	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/config"
	"github.com/cstevenson98/milo/pkg/debug"
	"github.com/cstevenson98/milo/pkg/ecs"
	"github.com/cstevenson98/milo/pkg/imgui"
	"github.com/cstevenson98/milo/pkg/input"
	"github.com/cstevenson98/milo/pkg/logger"
	"github.com/cstevenson98/milo/pkg/render"
	"github.com/cstevenson98/milo/pkg/state"
	"github.com/cstevenson98/milo/pkg/types"
	"github.com/cstevenson98/milo/pkg/ui"
)

// Engine manages the canvas, input, and game loop. It implements the
// ebiten.Game interface (Update, Draw, Layout). Games register states.State
// implementations against types.GameState values; the engine renders the active
// state's ECS world and drives its Update.
type Engine struct {
	cfg config.Settings

	canvasManager *canvas.Canvas
	inputCapturer *input.Input
	ui            types.UIManager

	running           bool
	currentGameState  types.GameState
	registeredStates  map[types.GameState]state.State
	currentState      state.State
	renderer          *render.Renderer
	stateLock         sync.Mutex
	screenWidth       float64
	screenHeight      float64
	gameStateProvider interface{}

	// Pending state switch. Requests made during a state's Update are recorded
	// here and applied at the start of the next frame, so a state is never torn
	// down while it is still running its own Update (see RequestStateChange).
	pendingState    types.GameState
	hasPendingState bool

	// Optional Dear ImGui context. Nil unless EnableImGui was called.
	imguiCtx *imgui.Context
}

// NewEngine creates a new game engine instance from cfg. Pass config.Default()
// for the engine's stock settings, or a customized config.Settings to override
// screen resolution, rendering quality, or debug-console behavior/appearance.
func NewEngine(cfg config.Settings) *Engine {
	return &Engine{
		cfg: cfg,
		canvasManager: canvas.NewCanvas(canvas.Config{
			PixelArtMode:        cfg.Rendering.PixelArtMode,
			PixelPerfectScaling: cfg.Rendering.PixelPerfectScaling,
		}),
		inputCapturer:    input.NewInput(),
		running:          false,
		registeredStates: make(map[types.GameState]state.State),
		screenWidth:      cfg.Screen.Width,
		screenHeight:     cfg.Screen.Height,
	}
}

// EnableImGui opts into Dear ImGui for this engine. Call before Initialize.
// Returns the engine for chaining. On WebAssembly ImGui is a silent no-op.
// States that want to draw windows should implement imgui.StateRenderer.
func (e *Engine) EnableImGui() *Engine {
	if e.imguiCtx == nil {
		e.imguiCtx = imgui.NewContext()
	}
	return e
}

// RegisterState registers a state for a specific game state value.
func (e *Engine) RegisterState(gs types.GameState, s state.State) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	e.registeredStates[gs] = s
	logger.Logger.Debugf("Registered state for game state: %s", gs.String())
}

// Initialize sets up the engine.
func (e *Engine) Initialize(canvasID string) error {
	logger.Logger.Debugf("Engine initializing")

	if err := e.canvasManager.Initialize(canvasID); err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return err
	}

	if err := e.inputCapturer.Initialize(); err != nil {
		logger.Logger.Errorf("Failed to initialize input: %s", err.Error())
		return err
	}

	// Apply this engine's config to the shared debug console singleton, so a
	// customized config.Settings actually takes effect instead of the
	// console's built-in defaults.
	debug.Console.Configure(debug.Config{
		Enabled:         e.cfg.Debug.Enabled,
		MaxMessages:     e.cfg.Debug.MaxMessages,
		MessageLifetime: e.cfg.Debug.MessageLifetime,
		ConsoleHeight:   e.cfg.Debug.ConsoleHeight,
		ScreenWidth:     e.cfg.Screen.Width,
		BackgroundColor: e.cfg.Debug.BackgroundColor,
		TextColor:       e.cfg.Debug.TextColor,
	})

	// Create the shared immediate-mode UI facade so states can draw text and
	// primitives without loading their own fonts or touching the canvas. On
	// failure we fall back to a no-op so UI calls stay safe.
	e.ui = types.NopUI
	uiCfg := ui.Config{
		CharacterSpacingReduction: e.cfg.Debug.CharacterSpacingReduction,
		UILineSpacing:             e.cfg.Rendering.UILineSpacing,
		TextLineSpacing:           e.cfg.Rendering.TextLineSpacing,
	}
	if uiFacade, err := ui.New(e.canvasManager, e.cfg.Debug.FontPath, e.screenWidth, e.screenHeight, uiCfg); err != nil {
		logger.Logger.Warnf("Failed to create UI facade: %s", err.Error())
	} else {
		e.ui = uiFacade
	}

	if e.imguiCtx != nil {
		if err := e.imguiCtx.Init(int(e.screenWidth), int(e.screenHeight)); err != nil {
			logger.Logger.Warnf("Failed to initialize ImGui: %s", err.Error())
			e.imguiCtx = nil
		} else {
			logger.Logger.Debugf("ImGui enabled")
		}
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

	// Apply any state switch requested during the previous frame before running
	// the current state's Update, so switches never happen mid-Update.
	e.applyPendingStateChange()

	e.stateLock.Lock()
	currentState := e.currentState
	e.stateLock.Unlock()

	if currentState != nil {
		// Refresh the per-world input resource so systems see this frame's input.
		if in := ecs.GetResource[components.Input](currentState.World()); in != nil {
			in.State = e.inputCapturer.GetInputState()
		}
		currentState.Update(deltaTime)
	}

	// ImGui must build its UI on the Update goroutine: Ebiten may run Draw on
	// another thread, and cimgui is not thread-safe / requires WithinFrameScope
	// between NewFrame and EndFrame on the same caller.
	if e.imguiCtx != nil {
		e.imguiCtx.SetScreenSize(int(e.screenWidth), int(e.screenHeight))
		e.imguiCtx.NewFrame()
		if currentState != nil {
			if r, ok := currentState.(imgui.StateRenderer); ok {
				r.RenderImGui(e.imguiCtx)
			}
		}
		e.imguiCtx.EndFrame()
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
	currentState := e.currentState
	renderer := e.renderer
	e.stateLock.Unlock()

	if currentState == nil {
		return
	}

	// Render the world (sprites, in layer + Order order).
	if renderer != nil {
		renderer.Draw(e.canvasManager)
	}

	// Render state-specific overlays (menus, HUD, debug console).
	if overlay, ok := currentState.(state.OverlayRenderer); ok {
		if err := overlay.DrawOverlays(); err != nil {
			logger.Logger.Tracef("Failed to render state overlays: %s", err.Error())
		}
	}

	// Blit the ImGui draw lists built during Update.
	if e.imguiCtx != nil {
		e.imguiCtx.Draw(screen)
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

	if e.currentState != nil {
		e.currentState.Exit()
		e.currentState = nil
	}

	if e.inputCapturer != nil {
		e.inputCapturer.Cleanup()
	}

	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager.
func (e *Engine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}

// RegisterGameStateProvider registers a game-defined global state provider that
// is passed to states via Deps.GameState.
func (e *Engine) RegisterGameStateProvider(provider interface{}) {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	e.gameStateProvider = provider
	logger.Logger.Debugf("Registered game state provider with engine")
}

// SetGameState changes the current game state and switches the active state
// immediately. Use this to activate the starting state before the game loop
// runs. States should not call this from within Update - use the injected
// RequestState callback instead, which defers the switch.
func (e *Engine) SetGameState(gs types.GameState) error {
	return e.applyGameState(gs)
}

// RequestStateChange asks the engine to switch to gs on the next frame. It is
// the callback injected into states via Deps, so a state can request a switch
// from within its own Update without being torn down mid-Update.
func (e *Engine) RequestStateChange(gs types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	if _, exists := e.registeredStates[gs]; !exists {
		return fmt.Errorf("no state registered for game state: %s", gs.String())
	}

	e.pendingState = gs
	e.hasPendingState = true
	logger.Logger.Debugf("Requested game state change to: %s", gs.String())
	return nil
}

// applyPendingStateChange performs a deferred state switch, if one was
// requested. Called by Update at the start of each frame.
func (e *Engine) applyPendingStateChange() {
	e.stateLock.Lock()
	if !e.hasPendingState {
		e.stateLock.Unlock()
		return
	}
	gs := e.pendingState
	e.hasPendingState = false
	e.stateLock.Unlock()

	if err := e.applyGameState(gs); err != nil {
		logger.Logger.Errorf("Failed to apply pending state change to %s: %s", gs.String(), err.Error())
	}
}

// applyGameState performs the actual state switch: tear down the current state,
// enter the new one, and build its renderer. Textures are lazy-loaded by the
// canvas on first draw, so there is no separate asset-preload step.
func (e *Engine) applyGameState(gs types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	registered, exists := e.registeredStates[gs]
	if !exists {
		return fmt.Errorf("no state registered for game state: %s", gs.String())
	}

	// Tear down the outgoing state.
	if e.currentState != nil {
		e.currentState.Exit()
		e.currentState = nil
		e.renderer = nil
	}

	// Enter the state with engine dependencies.
	deps := state.Deps{
		Input:            e.inputCapturer,
		UI:               e.ui,
		ScreenWidth:      e.screenWidth,
		ScreenHeight:     e.screenHeight,
		RequestState:     e.RequestStateChange,
		GameState:        e.gameStateProvider,
		Debug:            state.DebugConfig{Enabled: e.cfg.Debug.Enabled},
		DefaultFrameTime: e.cfg.Animation.DefaultFrameTime,
	}
	if err := registered.Enter(deps); err != nil {
		return fmt.Errorf("failed to enter state: %w", err)
	}

	e.currentState = registered
	e.renderer = render.NewRenderer(registered.World())
	e.currentGameState = gs
	logger.Logger.Debugf("Game state changed to: %s", gs.String())
	return nil
}

// GetGameState returns the current game state.
func (e *Engine) GetGameState() types.GameState {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()
	return e.currentGameState
}

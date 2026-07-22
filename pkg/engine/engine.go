package engine

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/input"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/render"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
	"github.com/cstevenson98/gowasm-engine/pkg/ui"
)

// Engine manages the canvas, input, and game loop. It implements the
// ebiten.Game interface (Update, Draw, Layout). Games register states.State
// implementations against types.GameState values; the engine renders the active
// state's ECS world and drives its Update.
type Engine struct {
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
}

// NewEngine creates a new game engine instance.
func NewEngine() *Engine {
	return &Engine{
		canvasManager:    canvas.NewCanvas(),
		inputCapturer:    input.NewInput(),
		running:          false,
		registeredStates: make(map[types.GameState]state.State),
		screenWidth:      config.Global.Screen.Width,
		screenHeight:     config.Global.Screen.Height,
	}
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

	// Create the shared immediate-mode UI facade so states can draw text and
	// primitives without loading their own fonts or touching the canvas. On
	// failure we fall back to a no-op so UI calls stay safe.
	e.ui = types.NopUI
	if uiFacade, err := ui.New(e.canvasManager, config.Global.Debug.FontPath, e.screenWidth, e.screenHeight); err != nil {
		logger.Logger.Warnf("Failed to create UI facade: %s", err.Error())
	} else {
		e.ui = uiFacade
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
		Input:        e.inputCapturer,
		UI:           e.ui,
		ScreenWidth:  e.screenWidth,
		ScreenHeight: e.screenHeight,
		RequestState: e.RequestStateChange,
		GameState:    e.gameStateProvider,
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

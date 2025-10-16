//go:build js

package engine

import (
	"sync"
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/gameobject"
	"github.com/conor/webgpu-triangle/internal/input"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Engine represents the game engine that manages the canvas and game loop
type Engine struct {
	canvasManager        canvas.CanvasManager
	inputCapturer        types.InputCapturer
	player               *gameobject.Player
	lastTime             float64
	textureLoaded        bool
	running              bool
	currentGameState     types.GameState
	gameStatePipelines   map[types.GameState][]types.PipelineType
	gameStateGameObjects map[types.GameState][]types.GameObject // GameObjects for each game state
	stateLock            sync.Mutex                             // Lock to prevent concurrent state changes
}

// NewEngine creates a new game engine instance
func NewEngine() *Engine {
	e := &Engine{
		canvasManager:        canvas.NewWebGPUCanvasManager(),
		inputCapturer:        input.NewUnifiedInput(),
		running:              false,
		gameStatePipelines:   make(map[types.GameState][]types.PipelineType),
		gameStateGameObjects: make(map[types.GameState][]types.GameObject),
	}

	// Initialize game state pipeline mappings and game objects
	e.initializeGameStates()

	return e
}

// initializeGameStates sets up the pipeline configurations and game objects for each game state
func (e *Engine) initializeGameStates() {
	// Screen dimensions (should match canvas size)
	const screenWidth = 800.0
	const screenHeight = 600.0

	// SPRITE state uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.SPRITE] = []types.PipelineType{
		types.TexturedPipeline,
	}

	// Create Player GameObject in the center of the screen
	playerSize := 128.0
	e.player = gameobject.NewPlayer(
		types.Vector2{
			X: (screenWidth - playerSize) / 2,  // Center X
			Y: (screenHeight - playerSize) / 2, // Center Y
		},
		types.Vector2{X: playerSize, Y: playerSize},
		200.0, // Movement speed: 200 pixels per second
	)

	println("DEBUG: Created Player at center of screen")

	// No other game objects for now
	e.gameStateGameObjects[types.SPRITE] = []types.GameObject{}

	// TRIANGLE state (kept for state switching)
	e.gameStatePipelines[types.TRIANGLE] = []types.PipelineType{
		types.TrianglePipeline,
	}

	e.gameStateGameObjects[types.TRIANGLE] = []types.GameObject{}
}

// Initialize sets up the engine with the specified canvas ID
func (e *Engine) Initialize(canvasID string) error {
	println("DEBUG: Engine initializing with canvas:", canvasID)

	err := e.canvasManager.Initialize(canvasID)
	if err != nil {
		println("DEBUG: Engine initialization failed:", err.Error())
		return err
	}

	err = e.SetGameState(types.SPRITE)
	if err != nil {
		println("DEBUG: Failed to set initial game state:", err.Error())
		return err
	}

	// Initialize input capturer
	err = e.inputCapturer.Initialize()
	if err != nil {
		println("DEBUG: Failed to initialize input:", err.Error())
		return err
	}

	println("DEBUG: Engine initialized successfully")
	return nil
}

// Start begins th e game loop
func (e *Engine) Start() {
	if e.running {
		println("DEBUG: Engine already running")
		return
	}

	e.running = true
	println("DEBUG: Engine starting render loop")

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
	currentState := e.currentGameState
	gameObjects := e.gameStateGameObjects[currentState]
	e.stateLock.Unlock()

	// Only update player in SPRITE state
	if currentState == types.SPRITE && e.player != nil {
		// Get input state and apply to player
		inputState := e.inputCapturer.GetInputState()
		e.player.HandleInput(inputState)

		// Update player mover (position)
		if mover := e.player.GetMover(); mover != nil {
			mover.Update(deltaTime)
		}

		// Update player sprite (animation)
		if sprite := e.player.GetSprite(); sprite != nil {
			sprite.Update(deltaTime)
		}

		// Update player game logic
		e.player.Update(deltaTime)
	}

	// Update other game objects
	for _, gameObject := range gameObjects {
		if mover := gameObject.GetMover(); mover != nil {
			mover.Update(deltaTime)
		}

		if sprite := gameObject.GetSprite(); sprite != nil {
			sprite.Update(deltaTime)
		}

		gameObject.Update(deltaTime)
	}

	// TODO: This should be done in a separate thread,
	// at game state change, with a black screen in between.
	// Load textures for sprites if needed
	e.loadSpriteTextures()
}

// Render draws the current frame
func (e *Engine) Render() {
	e.stateLock.Lock()
	currentState := e.currentGameState
	gameObjects := e.gameStateGameObjects[currentState]
	e.stateLock.Unlock()

	// Check if we have anything to render
	hasPlayer := currentState == types.SPRITE && e.player != nil
	hasObjects := len(gameObjects) > 0

	if hasPlayer || hasObjects {
		err := e.canvasManager.BeginBatch()
		if err != nil {
			println("ERROR: Failed to begin batch:", err.Error())
		}
	}

	// Render player first (in SPRITE state)
	if hasPlayer {
		var renderData types.SpriteRenderData
		if mover := e.player.GetMover(); mover != nil {
			renderData = e.player.GetSprite().GetSpriteRenderData(mover.GetPosition())
		} else {
			renderData = e.player.GetSprite().GetSpriteRenderData(types.Vector2{X: 0, Y: 0})
		}

		if renderData.Visible {
			err := e.canvasManager.DrawTexturedRect(
				renderData.TexturePath,
				renderData.Position,
				renderData.Size,
				renderData.UV,
			)
			if err != nil {
				// Texture might not be loaded yet
			}
		}
	}

	// Render other game objects
	for _, gameObject := range gameObjects {
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
			continue
		}
	}

	if hasPlayer || hasObjects {
		err := e.canvasManager.EndBatch()
		if err != nil {
			println("ERROR: Failed to end batch:", err.Error())
		}
	}

	e.canvasManager.Render()
}

// loadSpriteTextures loads textures for all game objects in the current game state
func (e *Engine) loadSpriteTextures() {
	e.stateLock.Lock()
	currentState := e.currentGameState
	gameObjects := e.gameStateGameObjects[currentState]
	e.stateLock.Unlock()

	// Load player texture
	if currentState == types.SPRITE && e.player != nil {
		pos := types.Vector2{X: 0, Y: 0}
		if mover := e.player.GetMover(); mover != nil {
			pos = mover.GetPosition()
		}
		renderData := e.player.GetSprite().GetSpriteRenderData(pos)
		e.canvasManager.LoadTexture(renderData.TexturePath)
	}

	// Load other game object textures
	for _, gameObject := range gameObjects {
		pos := types.Vector2{X: 0, Y: 0}
		if mover := gameObject.GetMover(); mover != nil {
			pos = mover.GetPosition()
		}
		renderData := gameObject.GetSprite().GetSpriteRenderData(pos)
		e.canvasManager.LoadTexture(renderData.TexturePath)
	}
}

// Stop stops the game loop
func (e *Engine) Stop() {
	e.running = false
	println("DEBUG: Engine stopped")
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

	e.currentGameState = state
	println("DEBUG: Game state changed to:", state.String())
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

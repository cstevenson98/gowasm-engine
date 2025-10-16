//go:build js

package engine

import (
	"math/rand"
	"sync"
	"syscall/js"
	"time"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/gameobject"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Engine represents the game engine that manages the canvas and game loop
type Engine struct {
	canvasManager        canvas.CanvasManager
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
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Screen dimensions (should match canvas size)
	const screenWidth = 800.0
	const screenHeight = 600.0

	// SPRITE state uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.SPRITE] = []types.PipelineType{
		types.TexturedPipeline,
	}

	// Create 5 Llama GameObjects
	gameObjects := make([]types.GameObject, 5)

	for i := 0; i < 5; i++ {
		// Random position
		randomX := rand.Float64() * screenWidth
		randomY := rand.Float64() * (screenHeight - 128) // Keep sprites on screen (accounting for sprite size)

		// Random speed (50-150 pixels per second)
		randomSpeed := 50.0 + rand.Float64()*100.0

		// Random sprite size (64-256 pixels)
		randomSize := 64.0 + rand.Float64()*192.0

		llama := gameobject.NewLlama(
			types.Vector2{X: randomX, Y: randomY},
			types.Vector2{X: randomSize, Y: randomSize},
			randomSpeed,
		)

		gameObjects[i] = llama

		println("DEBUG: Created Llama GameObject", i+1, "at position (", randomX, ",", randomY, ") with speed", randomSpeed)
	}

	e.gameStateGameObjects[types.SPRITE] = gameObjects

	e.gameStatePipelines[types.TRIANGLE] = []types.PipelineType{
		types.TrianglePipeline,
	}

	e.gameStateGameObjects[types.TRIANGLE] = []types.GameObject{} // No game objects for triangle state
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

	e.setupKeyboardHandlers()

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

	if len(gameObjects) > 0 {
		err := e.canvasManager.BeginBatch()
		if err != nil {
			println("ERROR: Failed to begin batch:", err.Error())
		}
	}

	// Render each game object's sprite
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

	if len(gameObjects) > 0 {
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

	for _, gameObject := range gameObjects {
		renderData := gameObject.GetSprite().GetSpriteRenderData(gameObject.GetMover().GetPosition())
		// Try to load the texture (will be skipped if already loaded)
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
	return e.canvasManager.Cleanup()
}

// GetCanvasManager returns the underlying canvas manager for advanced usage
func (e *Engine) GetCanvasManager() canvas.CanvasManager {
	return e.canvasManager
}

// SetGameState changes the current game state and updates the active pipelines
// This method is thread-safe and locks state transitions
func (e *Engine) SetGameState(state types.GameState) error {
	e.stateLock.Lock()
	defer e.stateLock.Unlock()

	pipelines, exists := e.gameStatePipelines[state]
	if !exists {
		return &EngineError{Message: "Game state not configured: " + state.String()}
	}

	// Update canvas manager with the pipelines for this state
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

// setupKeyboardHandlers sets up keyboard event listeners for state switching
func (e *Engine) setupKeyboardHandlers() {
	println("DEBUG: Setting up keyboard handlers")

	// Create a callback function that persists across calls
	keydownHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}

		event := args[0]
		key := event.Get("key").String()

		switch key {
		case "1":
			println("DEBUG: Key '1' pressed - switching to SPRITE state")
			err := e.SetGameState(types.SPRITE)
			if err != nil {
				println("ERROR: Failed to switch to SPRITE state:", err.Error())
			}
		case "2":
			println("DEBUG: Key '2' pressed - switching to TRIANGLE state")
			err := e.SetGameState(types.TRIANGLE)
			if err != nil {
				println("ERROR: Failed to switch to TRIANGLE state:", err.Error())
			}
		}

		return nil
	})

	// Add event listener to the document
	js.Global().Get("document").Call("addEventListener", "keydown", keydownHandler)

	println("DEBUG: Keyboard handlers set up - Press '1' for SPRITE, '2' for TRIANGLE")
}

// EngineError represents an error in the engine
type EngineError struct {
	Message string
}

func (e *EngineError) Error() string {
	return e.Message
}

//go:build js

package engine

import (
	"sync"
	"syscall/js"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/input"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/scene"
	"github.com/conor/webgpu-triangle/internal/types"
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
	currentScene       scene.Scene // Active scene managing game objects
	stateLock          sync.Mutex  // Lock to prevent concurrent state changes
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
		screenWidth:        config.Global.Screen.Width,
		screenHeight:       config.Global.Screen.Height,
	}

	// Initialize game state pipeline mappings
	e.initializeGameStates()

	return e
}

// initializeGameStates sets up the pipeline configurations for each game state
func (e *Engine) initializeGameStates() {
	// SPRITE state uses textured pipeline for sprite rendering
	e.gameStatePipelines[types.GAMEPLAY] = []types.PipelineType{
		types.TexturedPipeline,
	}
}

// createSceneForState creates the appropriate scene for a given game state
func (e *Engine) createSceneForState(state types.GameState) scene.Scene {
	switch state {
	case types.GAMEPLAY:
		return scene.NewGameplayScene(e.screenWidth, e.screenHeight, e.inputCapturer)
	default:
		logger.Logger.Warnf("No scene defined for game state: %s", state.String())
		return nil
	}
}

// Initialize sets up the engine with the specified canvas ID
func (e *Engine) Initialize(canvasID string) error {
	logger.Logger.Debugf("Engine initializing with canvas: %s", canvasID)

	err := e.canvasManager.Initialize(canvasID)
	if err != nil {
		logger.Logger.Errorf("Engine initialization failed: %s", err.Error())
		return err
	}

	err = e.SetGameState(types.GAMEPLAY)
	if err != nil {
		logger.Logger.Errorf("Failed to set initial game state: %s", err.Error())
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

	// Cleanup old scene
	if e.currentScene != nil {
		e.currentScene.Cleanup()
		e.currentScene = nil
	}

	// Create and initialize new scene for this state
	newScene := e.createSceneForState(state)
	if newScene != nil {
		err = newScene.Initialize()
		if err != nil {
			return &EngineError{Message: "Failed to initialize scene: " + err.Error()}
		}
		e.currentScene = newScene
		logger.Logger.Debugf("Initialized scene: %s for game state: %s", newScene.GetName(), state.String())
	}

	e.currentGameState = state
	logger.Logger.Debugf("Game state changed to: %s", state.String())
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

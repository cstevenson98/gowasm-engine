//go:build js

package scenes

import (
	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/gameobject"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GameplayScene represents the main gameplay scene with player and game objects
type GameplayScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// State change callback (injected by engine)
	stateChangeCallback func(state types.GameState) error

	// Player (managed separately for input handling)
	player *gameobject.Player

	// Game objects organized by layer
	layers map[pkscene.SceneLayer][]types.GameObject

	// Debug rendering
	debugFont         text.Font
	debugTextRenderer text.TextRenderer
	canvasManager     canvas.CanvasManager

	// Key press state tracking
	key1PressedLastFrame bool
	key2PressedLastFrame bool
}

// NewGameplayScene creates a new gameplay scene
func NewGameplayScene(screenWidth, screenHeight float64) *GameplayScene {
	return &GameplayScene{
		name:         "Gameplay",
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		layers:       make(map[pkscene.SceneLayer][]types.GameObject),
	}
}

// SetInputCapturer implements types.SceneInputProvider
func (s *GameplayScene) SetInputCapturer(inputCapturer types.InputCapturer) {
	s.inputCapturer = inputCapturer
}

// SetStateChangeCallback implements types.SceneStateChangeRequester
func (s *GameplayScene) SetStateChangeCallback(callback func(state types.GameState) error) {
	s.stateChangeCallback = callback
}

// SetCanvasManager sets the canvas manager for debug rendering
func (s *GameplayScene) SetCanvasManager(cm canvas.CanvasManager) {
	s.canvasManager = cm
}

// GetRequiredAssets implements types.SceneAssetProvider
func (s *GameplayScene) GetRequiredAssets() types.SceneAssets {
	return types.SceneAssets{
		TexturePaths: []string{
			"art/test-background.png",
			config.Global.Player.TexturePath,
		},
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	}
}

// InitializeDebugConsole initializes the debug console font and text renderer
func (s *GameplayScene) InitializeDebugConsole() error {
	if !config.Global.Debug.Enabled {
		return nil
	}

	logger.Logger.Debugf("Initializing debug console for %s scene", s.name)

	// Create and load font metadata
	s.debugFont = text.NewSpriteFont()
	err := s.debugFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load debug font: %s", err)
		return err
	}

	// Create text renderer (texture will be loaded by engine's loadSpriteTextures)
	s.debugTextRenderer = text.NewTextRenderer(s.canvasManager)

	logger.Logger.Debugf("Debug console initialized successfully")
	// Post a welcome message after a short delay to allow texture loading
	debug.Console.PostMessage("System", "Debug console ready")

	return nil
}

// Initialize sets up the gameplay scene and creates game objects
func (s *GameplayScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.name)

	// Initialize layer slices
	s.layers[pkscene.BACKGROUND] = []types.GameObject{}
	s.layers[pkscene.ENTITIES] = []types.GameObject{}
	s.layers[pkscene.UI] = []types.GameObject{}

	// Create background (BACKGROUND layer)
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.screenWidth, Y: s.screenHeight},
		"art/test-background.png",
	)
	s.AddGameObject(pkscene.BACKGROUND, background)
	logger.Logger.Debugf("Created Background in %s scene", s.name)

	// Create player in the center of the screen (ENTITIES layer)
	spawnX, spawnY := config.GetPlayerSpawnPosition()
	s.player = gameobject.NewPlayer(
		types.Vector2{X: spawnX, Y: spawnY},
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)

	logger.Logger.Debugf("Created Player at center of screen in %s scene", s.name)

	return nil
}

// Update updates all game objects in the scene
func (s *GameplayScene) Update(deltaTime float64) {
	// Update player with input
	if s.player != nil && s.inputCapturer != nil {
		// Get input state and apply to player
		inputState := s.inputCapturer.GetInputState()
		s.player.HandleInput(inputState)

		// Handle scene switching: Key 1 switches to gameplay (no-op, already in gameplay), Key 2 switches to battle scene
		if inputState.Key2Pressed && !s.key2PressedLastFrame && s.stateChangeCallback != nil {
			logger.Logger.Debugf("Key 2 pressed: switching to battle scene")
			err := s.stateChangeCallback(types.BATTLE)
			if err != nil {
				logger.Logger.Errorf("Failed to switch to battle scene: %s", err.Error())
			}
			// Return early - scene may have been cleaned up during state change
			s.key1PressedLastFrame = inputState.Key1Pressed
			s.key2PressedLastFrame = inputState.Key2Pressed
			return
		}
		s.key1PressedLastFrame = inputState.Key1Pressed
		s.key2PressedLastFrame = inputState.Key2Pressed

		// Re-check player exists (may have been cleaned up during state change)
		if s.player == nil {
			return
		}

		// Update player mover (position)
		if mover := s.player.GetMover(); mover != nil {
			mover.Update(deltaTime)
		}

		// Update player sprite (animation)
		if sprite := s.player.GetSprite(); sprite != nil {
			sprite.Update(deltaTime)
		}

		// Update player game logic
		s.player.Update(deltaTime)
	}

	// Update all game objects in all layers
	for _, layer := range []pkscene.SceneLayer{pkscene.BACKGROUND, pkscene.ENTITIES, pkscene.UI} {
		for _, gameObject := range s.layers[layer] {
			if mover := gameObject.GetMover(); mover != nil {
				mover.Update(deltaTime)
			}

			if sprite := gameObject.GetSprite(); sprite != nil {
				sprite.Update(deltaTime)
			}

			gameObject.Update(deltaTime)
		}
	}

	// Update debug console
	if config.Global.Debug.Enabled {
		debug.Console.Update(deltaTime)
	}
}

// RenderDebugConsole renders the debug console UI
func (s *GameplayScene) RenderDebugConsole() error {
	if !config.Global.Debug.Enabled || s.debugFont == nil || s.debugTextRenderer == nil {
		return nil
	}

	return debug.Console.Render(s.canvasManager, s.debugTextRenderer, s.debugFont)
}

// GetRenderables returns all game objects in the correct render order
func (s *GameplayScene) GetRenderables() []types.GameObject {
	var result []types.GameObject

	// Render layers in order: BACKGROUND → ENTITIES → UI
	for _, layer := range []pkscene.SceneLayer{pkscene.BACKGROUND, pkscene.ENTITIES, pkscene.UI} {
		// Add player to ENTITIES layer during rendering
		if layer == pkscene.ENTITIES && s.player != nil {
			result = append(result, s.player)
		}

		// Add other game objects in this layer
		result = append(result, s.layers[layer]...)
	}

	return result
}

// Cleanup releases scene resources
func (s *GameplayScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.name)

	// Clear player reference
	s.player = nil

	// Clear all layers
	for layer := range s.layers {
		s.layers[layer] = nil
	}
	s.layers = make(map[pkscene.SceneLayer][]types.GameObject)
}

// GetName returns the scene identifier
func (s *GameplayScene) GetName() string {
	return s.name
}

// AddGameObject adds a game object to the specified layer
func (s *GameplayScene) AddGameObject(layer pkscene.SceneLayer, obj types.GameObject) {
	s.layers[layer] = append(s.layers[layer], obj)
	logger.Logger.Debugf("Added GameObject to %s layer in %s scene", layer.String(), s.name)
}

// RemoveGameObject removes a game object from the specified layer
func (s *GameplayScene) RemoveGameObject(layer pkscene.SceneLayer, obj types.GameObject) {
	objects := s.layers[layer]
	for i, o := range objects {
		if o == obj {
			s.layers[layer] = append(objects[:i], objects[i+1:]...)
			logger.Logger.Debugf("Removed GameObject from %s layer in %s scene", layer.String(), s.name)
			return
		}
	}
}

// GetPlayer returns the player object (for special access if needed)
func (s *GameplayScene) GetPlayer() *gameobject.Player {
	return s.player
}

// GetDebugFont returns the debug font (for texture loading)
func (s *GameplayScene) GetDebugFont() text.Font {
	return s.debugFont
}

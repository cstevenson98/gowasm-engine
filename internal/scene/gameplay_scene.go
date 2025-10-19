//go:build js

package scene

import (
	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/gameobject"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/types"
)

// GameplayScene represents the main gameplay scene with player and game objects
type GameplayScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// Player (managed separately for input handling)
	player *gameobject.Player

	// Game objects organized by layer
	layers map[SceneLayer][]types.GameObject
}

// NewGameplayScene creates a new gameplay scene
func NewGameplayScene(screenWidth, screenHeight float64, inputCapturer types.InputCapturer) *GameplayScene {
	return &GameplayScene{
		name:          "Gameplay",
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		inputCapturer: inputCapturer,
		layers:        make(map[SceneLayer][]types.GameObject),
	}
}

// Initialize sets up the gameplay scene and creates game objects
func (s *GameplayScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.name)

	// Initialize layer slices
	s.layers[BACKGROUND] = []types.GameObject{}
	s.layers[ENTITIES] = []types.GameObject{}
	s.layers[UI] = []types.GameObject{}

	// Create background (BACKGROUND layer)
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.screenWidth, Y: s.screenHeight},
		"art/test-background.png",
	)
	s.AddGameObject(BACKGROUND, background)
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
	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
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
}

// GetRenderables returns all game objects in the correct render order
func (s *GameplayScene) GetRenderables() []types.GameObject {
	var result []types.GameObject

	// Render layers in order: BACKGROUND → ENTITIES → UI
	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
		// Add player to ENTITIES layer during rendering
		if layer == ENTITIES && s.player != nil {
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
	s.layers = make(map[SceneLayer][]types.GameObject)
}

// GetName returns the scene identifier
func (s *GameplayScene) GetName() string {
	return s.name
}

// AddGameObject adds a game object to the specified layer
func (s *GameplayScene) AddGameObject(layer SceneLayer, obj types.GameObject) {
	s.layers[layer] = append(s.layers[layer], obj)
	logger.Logger.Debugf("Added GameObject to %s layer in %s scene", layer.String(), s.name)
}

// RemoveGameObject removes a game object from the specified layer
func (s *GameplayScene) RemoveGameObject(layer SceneLayer, obj types.GameObject) {
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

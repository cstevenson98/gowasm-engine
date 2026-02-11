//go:build js

package scenes

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
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

	// Game state manager (injected by engine)
	gameStateManager interface{} // Cast to *gamestate.GameStateManager

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
	mPressedLastFrame    bool // M key for player menu

	// Saved state for persistence (optional, used when scene implements SceneStateful)
	savedPlayerPosition *types.Vector2     // Player position to restore
	savedPlayerState    *types.ObjectState // Full player state to restore
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

// SetStateChangeCallback implements types.SceneChangeRequester
func (s *GameplayScene) SetStateChangeCallback(callback func(state types.GameState) error) {
	s.stateChangeCallback = callback
}

// SetGameState implements types.SceneGameStateUser
func (s *GameplayScene) SetGameState(gameState interface{}) {
	// Cast to the game's state manager type
	s.gameStateManager = gameState
	logger.Logger.Debugf("Gameplay scene received game state manager")
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

	// Create player - use saved position from global game state if available,
	// otherwise use scene-level saved position (for scene switching),
	// otherwise use spawn position
	var playerPos types.Vector2

	// First, check if we have a global game state with saved position (from load game)
	// The game state will have a non-zero Timestamp if it was loaded from a save
	if s.gameStateManager != nil {
		if manager, ok := s.gameStateManager.(*gamestate.GameStateManager); ok {
			globalState := manager.GetState()
			if globalState != nil && globalState.Timestamp > 0 {
				// Use position from loaded game state (Timestamp > 0 means it was loaded from a save)
				playerPos = globalState.PlayerPosition
				logger.Logger.Debugf("Creating Player at loaded position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
			} else if s.savedPlayerPosition != nil {
				// Fallback to scene-level saved position (for scene switching)
				playerPos = *s.savedPlayerPosition
				logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
			} else {
				// Default to spawn position
				spawnX, spawnY := config.GetPlayerSpawnPosition()
				playerPos = types.Vector2{X: spawnX, Y: spawnY}
				logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
			}
		} else if s.savedPlayerPosition != nil {
			// Fallback: scene-level saved position
			playerPos = *s.savedPlayerPosition
			logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
		} else {
			// Default to spawn position
			spawnX, spawnY := config.GetPlayerSpawnPosition()
			playerPos = types.Vector2{X: spawnX, Y: spawnY}
			logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
		}
	} else if s.savedPlayerPosition != nil {
		// Fallback: scene-level saved position
		playerPos = *s.savedPlayerPosition
		logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
	} else {
		// Default to spawn position
		spawnX, spawnY := config.GetPlayerSpawnPosition()
		playerPos = types.Vector2{X: spawnX, Y: spawnY}
		logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.name)
	}

	s.player = gameobject.NewPlayer(
		playerPos,
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)

	// Update game state manager with player reference (player is part of game state)
	if s.gameStateManager != nil {
		if manager, ok := s.gameStateManager.(*gamestate.GameStateManager); ok {
			manager.SetPlayer(s.player)
			logger.Logger.Debugf("Updated game state manager with player reference")
		}
	}

	// Note: Full state restoration happens in RestoreState() after Initialize() completes
	// This ensures the player is fully created before we restore state

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

		// Handle player menu (M key)
		if inputState.MPressed && !s.mPressedLastFrame && s.stateChangeCallback != nil {
			logger.Logger.Debugf("M key pressed: opening player menu")
			err := s.stateChangeCallback(types.PLAYER_MENU)
			if err != nil {
				logger.Logger.Errorf("Failed to switch to player menu: %s", err.Error())
			}
			// Return early - scene may have been cleaned up during state change
			s.mPressedLastFrame = inputState.MPressed
			return
		}
		s.mPressedLastFrame = inputState.MPressed

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

// SaveState implements types.SceneStateful
// Saves the current player position and state before cleanup
func (s *GameplayScene) SaveState() {
	if s.player != nil {
		// Save player position
		if mover := s.player.GetMover(); mover != nil {
			pos := mover.GetPosition()
			s.savedPlayerPosition = &pos
			logger.Logger.Debugf("Saved player position: (%.2f, %.2f)", pos.X, pos.Y)
		}

		// Save full player state
		playerState := s.player.GetState()
		if playerState != nil {
			savedState := types.CopyObjectState(*playerState)
			s.savedPlayerState = &savedState
			logger.Logger.Debugf("Saved player state for %s scene", s.name)
		}
	}
}

// RestoreState implements types.SceneStateful
// Restores the previously saved player position and state after initialization
func (s *GameplayScene) RestoreState() {
	if s.player == nil {
		return
	}

	// Player position is restored during Initialize() using savedPlayerPosition
	// Here we restore the full state if it was saved
	if s.savedPlayerState != nil {
		s.player.SetState(*s.savedPlayerState)
		logger.Logger.Debugf("Restored player state in %s scene", s.name)
	} else if s.savedPlayerPosition != nil {
		// If we have position but no full state, at least restore position
		if mover := s.player.GetMover(); mover != nil {
			mover.SetPosition(*s.savedPlayerPosition)
			logger.Logger.Debugf("Restored player position to: (%.2f, %.2f)", s.savedPlayerPosition.X, s.savedPlayerPosition.Y)
		}
	}
}

// handleSaveGame handles saving the current game state
func (s *GameplayScene) handleSaveGame() {
	if s.gameStateManager == nil {
		logger.Logger.Warnf("Cannot save: game state manager not available")
		return
	}

	// Cast to the game's state manager type
	manager, ok := s.gameStateManager.(*gamestate.GameStateManager)
	if !ok {
		logger.Logger.Warnf("Cannot save: invalid game state manager type")
		return
	}

	if s.player == nil {
		logger.Logger.Warnf("Cannot save: player not available")
		return
	}

	// Get current game state
	currentState := manager.GetState()
	if currentState == nil {
		logger.Logger.Warnf("Cannot save: no game state exists (create a new game first)")
		return
	}

	// Collect player position
	var playerPos types.Vector2
	if mover := s.player.GetMover(); mover != nil {
		playerPos = mover.GetPosition()
	}

	// Collect player stats
	var playerStats gamestate.PlayerStats
	if stats := s.player.GetStats(); stats != nil {
		playerStats = gamestate.PlayerStats{
			Level:      1, // TODO: get from player when leveling is implemented
			HP:         stats.HP,
			MaxHP:      stats.MaxHP,
			Experience: 0, // TODO: get from player when XP is implemented
		}
	} else {
		// Default stats if player doesn't have stats component
		playerStats = gamestate.PlayerStats{
			Level:      1,
			HP:         config.Global.Battle.PlayerHP,
			MaxHP:      config.Global.Battle.PlayerMaxHP,
			Experience: 0,
		}
	}

	// Update game state
	currentState.PlayerPosition = playerPos
	currentState.PlayerStats = playerStats
	// StoryState can be updated separately if needed

	// Save to localStorage
	saveKey, err := manager.SaveCurrentGame()
	if err != nil {
		logger.Logger.Errorf("Failed to save game: %s", err.Error())
		debug.Console.PostMessage("System", fmt.Sprintf("Save failed: %s", err.Error()))
	} else {
		logger.Logger.Infof("Game saved successfully: %s", saveKey)
		debug.Console.PostMessage("System", "Game saved successfully")
	}
}

// GetDebugFont returns the debug font (for texture loading)
func (s *GameplayScene) GetDebugFont() text.Font {
	return s.debugFont
}

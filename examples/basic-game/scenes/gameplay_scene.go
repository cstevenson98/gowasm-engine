//go:build js

package scenes

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/gameobject"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GameplayScene represents the main gameplay scene with player and game objects.
// It embeds BaseScene to inherit all common scene functionality.
type GameplayScene struct {
	*pkscene.BaseScene

	// Gameplay-specific fields
	player *gameobject.Player

	// Key press state tracking
	key1PressedLastFrame bool
	key2PressedLastFrame bool
	mPressedLastFrame    bool // M key for player menu
}

// NewGameplayScene creates a new gameplay scene
func NewGameplayScene(screenWidth, screenHeight float64) *GameplayScene {
	baseScene := pkscene.NewBaseScene("Gameplay", screenWidth, screenHeight)
	
	// Set required assets
	baseScene.SetRequiredAssets(types.SceneAssets{
		TexturePaths: []string{
			"art/test-background.png",
			config.Global.Player.TexturePath,
		},
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	})
	
	return &GameplayScene{
		BaseScene: baseScene,
	}
}

// All interface implementations (SetInputCapturer, SetStateChangeCallback, SetGameState, 
// SetCanvasManager, GetRequiredAssets) are inherited from BaseScene

// Initialize sets up the gameplay scene and creates game objects (overrides BaseScene.Initialize)
func (s *GameplayScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.GetName())

	// Call base initialization (sets up layers)
	if err := s.BaseScene.Initialize(); err != nil {
		return err
	}

	// Create background using BaseScene helper
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.GetScreenWidth(), Y: s.GetScreenHeight()},
		"art/test-background.png",
	)
	s.AddBackground(background)
	logger.Logger.Debugf("Created Background in %s scene", s.GetName())

	// Create player - use saved position from global game state if available,
	// otherwise use scene-level saved position (for scene switching),
	// otherwise use spawn position
	var playerPos types.Vector2

	// First, check if we have a global game state with saved position (from load game)
	if gameState := s.GetGameState(); gameState != nil {
		if manager, ok := gameState.(*gamestate.GameStateManager); ok {
			globalState := manager.GetState()
			if globalState != nil && globalState.Timestamp > 0 {
				// Use position from loaded game state
				playerPos = globalState.PlayerPosition
				logger.Logger.Debugf("Creating Player at loaded position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
			} else if savedPos, ok := s.GetSavedState()["playerPosition"].(types.Vector2); ok {
				// Fallback to scene-level saved position (for scene switching)
				playerPos = savedPos
				logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
			} else {
				// Default to spawn position
				spawnX, spawnY := config.GetPlayerSpawnPosition()
				playerPos = types.Vector2{X: spawnX, Y: spawnY}
				logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
			}
		} else if savedPos, ok := s.GetSavedState()["playerPosition"].(types.Vector2); ok {
			// Fallback: scene-level saved position
			playerPos = savedPos
			logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
		} else {
			// Default to spawn position
			spawnX, spawnY := config.GetPlayerSpawnPosition()
			playerPos = types.Vector2{X: spawnX, Y: spawnY}
			logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
		}
	} else if savedPos, ok := s.GetSavedState()["playerPosition"].(types.Vector2); ok {
		// Fallback: scene-level saved position
		playerPos = savedPos
		logger.Logger.Debugf("Creating Player at saved scene position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
	} else {
		// Default to spawn position
		spawnX, spawnY := config.GetPlayerSpawnPosition()
		playerPos = types.Vector2{X: spawnX, Y: spawnY}
		logger.Logger.Debugf("Creating Player at spawn position (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())
	}

	s.player = gameobject.NewPlayer(
		playerPos,
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)

	// Add player to ENTITIES layer using BaseScene helper
	s.AddEntity(s.player)

	// Update game state manager with player reference (player is part of game state)
	if gameState := s.GetGameState(); gameState != nil {
		if manager, ok := gameState.(*gamestate.GameStateManager); ok {
			manager.SetPlayer(s.player)
			logger.Logger.Debugf("Updated game state manager with player reference")
		}
	}

	// Note: Full state restoration happens in RestoreState() after Initialize() completes
	// This ensures the player is fully created before we restore state

	return nil
}

// Update updates all game objects in the scene (overrides BaseScene.Update)
func (s *GameplayScene) Update(deltaTime float64) {
	// Update player with input
	if s.player != nil {
		// Get input state using inherited method
		inputState := s.GetInputState()
		s.player.HandleInput(inputState)

		// Handle scene switching: Key 2 switches to battle scene
		if inputState.Key2Pressed && !s.key2PressedLastFrame {
			logger.Logger.Debugf("Key 2 pressed: switching to battle scene")
			err := s.RequestStateChange(types.BATTLE)
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
		if inputState.MPressed && !s.mPressedLastFrame {
			logger.Logger.Debugf("M key pressed: opening player menu")
			err := s.RequestStateChange(types.PLAYER_MENU)
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

	// Update all game objects in all layers using BaseScene method
	s.BaseScene.Update(deltaTime)

	// Update debug console
	if config.Global.Debug.Enabled {
		debug.Console.Update(deltaTime)
	}
}

// RenderOverlays renders debug console and other overlays (overrides BaseScene.RenderOverlays)
func (s *GameplayScene) RenderOverlays() error {
	// Render debug console (inherited from BaseScene)
	return s.BaseScene.RenderOverlays()
}

// Cleanup overrides BaseScene.Cleanup to also clear player reference
func (s *GameplayScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.GetName())
	
	// Clear player reference
	s.player = nil
	
	// Call base cleanup (clears all layers)
	s.BaseScene.Cleanup()
}

// SaveState implements types.SceneStateful (overrides BaseScene.SaveState)
// Saves the current player position and state before cleanup
func (s *GameplayScene) SaveState() {
	if s.player != nil {
		// Save player position in BaseScene's saved state map
		if mover := s.player.GetMover(); mover != nil {
			pos := mover.GetPosition()
			s.GetSavedState()["playerPosition"] = pos
			logger.Logger.Debugf("Saved player position: (%.2f, %.2f)", pos.X, pos.Y)
		}

		// Save full player state
		playerState := s.player.GetState()
		if playerState != nil {
			savedState := types.CopyObjectState(*playerState)
			s.GetSavedState()["playerState"] = savedState
			logger.Logger.Debugf("Saved player state for %s scene", s.GetName())
		}
	}
}

// RestoreState implements types.SceneStateful (overrides BaseScene.RestoreState)
// Restores the previously saved player position and state after initialization
func (s *GameplayScene) RestoreState() {
	if s.player == nil {
		return
	}

	// Player position is already restored during Initialize() using saved state
	// Here we restore the full state if it was saved
	if savedPlayerState, ok := s.GetSavedState()["playerState"].(types.ObjectState); ok {
		s.player.SetState(savedPlayerState)
		logger.Logger.Debugf("Restored player state in %s scene", s.GetName())
	} else if savedPos, ok := s.GetSavedState()["playerPosition"].(types.Vector2); ok {
		// If we have position but no full state, at least restore position
		if mover := s.player.GetMover(); mover != nil {
			mover.SetPosition(savedPos)
			logger.Logger.Debugf("Restored player position to: (%.2f, %.2f)", savedPos.X, savedPos.Y)
		}
	}
}

// handleSaveGame handles saving the current game state
func (s *GameplayScene) handleSaveGame() {
	gameState := s.GetGameState()
	if gameState == nil {
		logger.Logger.Warnf("Cannot save: game state manager not available")
		return
	}

	// Cast to the game's state manager type
	manager, ok := gameState.(*gamestate.GameStateManager)
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

// GetDebugFont is inherited from BaseScene

package scenes

import (
	"fmt"

	"example.com/basic-game/game/entities"
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
	player *entities.Player
}

// NewGameplayScene creates a new gameplay scene
func NewGameplayScene(screenWidth, screenHeight float64) *GameplayScene {
	baseScene := pkscene.NewBaseScene("Gameplay", screenWidth, screenHeight)

	// Set required assets
	baseScene.SetRequiredAssets(types.SceneAssets{
		TexturePaths: []string{
			"assets/art/test-background.png",
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
// GetRequiredAssets) are inherited from BaseScene

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
		"assets/art/test-background.png",
	)
	s.AddBackground(background)
	logger.Logger.Debugf("Created Background in %s scene", s.GetName())

	// Create the player at the position held in the global game state (set on
	// new game / load, and kept up to date when leaving this scene), falling
	// back to the configured spawn point.
	playerPos := s.resolvePlayerPosition()
	logger.Logger.Debugf("Creating Player at (%.2f, %.2f) in %s scene", playerPos.X, playerPos.Y, s.GetName())

	s.player = entities.NewPlayer(
		playerPos,
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)

	// Add player to ENTITIES layer using BaseScene helper
	s.AddEntity(s.player)

	// Update game state manager with player reference (player is part of game state)
	if manager := s.stateManager(); manager != nil {
		manager.SetPlayer(s.player)
		logger.Logger.Debugf("Updated game state manager with player reference")
	}

	return nil
}

// stateManager returns the game's state manager, or nil if unavailable.
func (s *GameplayScene) stateManager() *gamestate.GameStateManager {
	if gameState := s.GetGameState(); gameState != nil {
		if manager, ok := gameState.(*gamestate.GameStateManager); ok {
			return manager
		}
	}
	return nil
}

// resolvePlayerPosition returns where the player should spawn: the position in
// the global game state if one exists, otherwise the configured spawn point.
func (s *GameplayScene) resolvePlayerPosition() types.Vector2 {
	if manager := s.stateManager(); manager != nil {
		if state := manager.GetState(); state != nil {
			return state.PlayerPosition
		}
	}
	spawnX, spawnY := config.GetPlayerSpawnPosition()
	return types.Vector2{X: spawnX, Y: spawnY}
}

// Update updates all game objects in the scene (overrides BaseScene.Update)
func (s *GameplayScene) Update(deltaTime float64) {
	inputState := s.GetInputState()

	// Feed input to the player before objects are advanced.
	if s.player != nil {
		s.player.HandleInput(inputState)
	}

	// Scene switching (edge-detected using the input snapshot's last-frame flags).
	// The engine defers the switch to the next frame, so it is safe to continue.
	if inputState.Key2Pressed && !inputState.Key2PressedLastFrame {
		logger.Logger.Debugf("Key 2 pressed: switching to battle scene")
		if err := s.RequestStateChange(types.BATTLE); err != nil {
			logger.Logger.Errorf("Failed to switch to battle scene: %s", err.Error())
		}
	} else if inputState.MPressed && !inputState.MPressedLastFrame {
		logger.Logger.Debugf("M key pressed: opening player menu")
		if err := s.RequestStateChange(types.PLAYER_MENU); err != nil {
			logger.Logger.Errorf("Failed to switch to player menu: %s", err.Error())
		}
	}

	// Advance every object in the scene (player included: BaseGameObject.Update
	// moves the mover and animates the sprite) and drive the debug console.
	s.BaseScene.Update(deltaTime)
}

// RenderOverlays renders debug console and other overlays (overrides BaseScene.RenderOverlays)
func (s *GameplayScene) RenderOverlays() error {
	// Render debug console (inherited from BaseScene)
	return s.BaseScene.RenderOverlays()
}

// Cleanup overrides BaseScene.Cleanup. Before tearing down, it writes the
// player's live position back into the global game state so that returning to
// gameplay (e.g. after the player menu) resumes from where the player was.
func (s *GameplayScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.GetName())

	if s.player != nil {
		if manager := s.stateManager(); manager != nil {
			if mover := s.player.GetMover(); mover != nil {
				manager.UpdatePlayerPosition(mover.GetPosition())
			}
		}
	}

	// Clear player reference
	s.player = nil

	// Call base cleanup (clears all layers)
	s.BaseScene.Cleanup()
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

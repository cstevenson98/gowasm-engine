package scenes

import (
	"fmt"

	"example.com/basic-game/game/entities"
	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// PlayerMenuScene represents the player menu scene accessible from gameplay.
// It embeds BaseScene to inherit all common scene functionality.
type PlayerMenuScene struct {
	*pkscene.BaseScene

	// Player reference (passed from gameplay scene)
	player *entities.Player

	// Menu system
	menuSystem *PlayerMenuSystem
}

// NewPlayerMenuScene creates a new player menu scene
func NewPlayerMenuScene(screenWidth, screenHeight float64) *PlayerMenuScene {
	baseScene := pkscene.NewBaseScene("PlayerMenu", screenWidth, screenHeight)

	// Declare required assets. The font texture is loaded automatically from
	// FontPaths, so it doesn't need to be listed under TexturePaths.
	baseScene.SetRequiredAssets(types.SceneAssets{
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	})

	return &PlayerMenuScene{
		BaseScene: baseScene,
	}
}

// All interface implementations (SetInputCapturer, SetStateChangeCallback, SetGameState)
// are inherited from BaseScene

// updatePlayerReference updates the player reference from the game state manager
// Only updates if the player reference has changed to avoid excessive logging
func (s *PlayerMenuScene) updatePlayerReference() {
	gameState := s.GetGameState()
	if gameState == nil {
		if s.player != nil {
			s.player = nil
		}
		return
	}

	// Get player from game state manager
	manager, ok := gameState.(*gamestate.GameStateManager)
	if !ok {
		return
	}
	player := manager.GetPlayer()
	if player == nil {
		if s.player != nil {
			s.player = nil
		}
		return
	}

	// Cast to the game's player type
	if p, ok := player.(*entities.Player); ok {
		// Only log if player changed
		if s.player != p {
			s.player = p
			logger.Logger.Debugf("Player menu scene retrieved player reference from game state")
		}
	} else {
		if s.player != nil {
			s.player = nil
			logger.Logger.Warnf("Player menu scene: invalid player type in game state")
		}
	}
}

// Initialize sets up the player menu scene (overrides BaseScene.Initialize)
func (s *PlayerMenuScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.GetName())

	// Call base initialization (sets up layers)
	if err := s.BaseScene.Initialize(); err != nil {
		return err
	}

	// Initialize menu system
	s.menuSystem = NewPlayerMenuSystem(s.GetScreenWidth(), s.GetScreenHeight())
	s.menuSystem.Initialize()

	return nil
}

// Update updates the player menu scene
func (s *PlayerMenuScene) Update(deltaTime float64) {
	// Update player reference from game state (player is part of game state)
	s.updatePlayerReference()

	inputState := s.GetInputState()

	// Handle menu close (M key). The engine defers the switch to the next frame.
	if inputState.MPressed && !inputState.MPressedLastFrame {
		logger.Logger.Debugf("M key pressed: closing player menu")
		if err := s.RequestStateChange(types.GAMEPLAY); err != nil {
			logger.Logger.Errorf("Failed to switch back to gameplay: %s", err.Error())
		}
		return
	}

	// Handle menu navigation
	menu := s.menuSystem.playerMenu

	// Navigation
	if inputState.UpPressed && !inputState.UpPressedLastFrame {
		menu.selectedIndex--
		if menu.selectedIndex < 0 {
			menu.selectedIndex = len(menu.options) - 1
		}
	}
	if inputState.DownPressed && !inputState.DownPressedLastFrame {
		menu.selectedIndex++
		if menu.selectedIndex >= len(menu.options) {
			menu.selectedIndex = 0
		}
	}

	// Selection
	if inputState.EnterPressed && !inputState.EnterPressedLastFrame {
		selected := menu.options[menu.selectedIndex]
		if selected == "Save Game" {
			s.handleSaveGame()
		}
	}

	// Drive shared behaviour (debug console toggle + aging).
	s.BaseScene.Update(deltaTime)
}

// handleSaveGame handles saving the current game state and shows browser alert
func (s *PlayerMenuScene) handleSaveGame() {
	gameState := s.GetGameState()
	if gameState == nil {
		s.showAlert("Save failed: Game state manager not available")
		return
	}

	manager, ok := gameState.(*gamestate.GameStateManager)
	if !ok {
		s.showAlert("Save failed: Invalid game state manager")
		return
	}

	if s.player == nil {
		s.showAlert("Save failed: Player not available")
		return
	}

	// Get current game state
	currentState := manager.GetState()
	if currentState == nil {
		s.showAlert("Save failed: No game state exists (create a new game first)")
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

	// Update game state with current player data (thread-safe, holds lock)
	manager.UpdateStateFromPlayer(playerPos, playerStats)
	logger.Logger.Debugf("Updated game state before save - Position: (%.2f, %.2f)", playerPos.X, playerPos.Y)

	// Save to localStorage
	saveKey, err := gameState.(*gamestate.GameStateManager).SaveCurrentGame()
	if err != nil {
		logger.Logger.Errorf("Failed to save game: %s", err.Error())
		s.showAlert(fmt.Sprintf("Save failed: %s", err.Error()))
	} else {
		logger.Logger.Infof("Game saved successfully: %s", saveKey)
		s.showAlert("Game saved successfully!")
		debug.Console.PostMessage("System", "Game saved successfully")
	}
}

// showAlert shows an alert message (logs to console and debug overlay)
func (s *PlayerMenuScene) showAlert(message string) {
	logger.Logger.Infof("Alert: %s", message)
	debug.Console.PostMessage("Alert", message)
}

// RenderOverlays implements types.SceneOverlayRenderer
func (s *PlayerMenuScene) RenderOverlays() error {
	// Render player info on left
	if err := s.renderPlayerInfo(); err != nil {
		logger.Logger.Tracef("Failed to render player info: %s", err)
	}

	// Render menu on right
	if err := s.renderMenu(); err != nil {
		logger.Logger.Tracef("Failed to render menu: %s", err)
	}

	return nil
}

// renderPlayerInfo renders player information on the left side
func (s *PlayerMenuScene) renderPlayerInfo() error {
	if s.player == nil {
		return nil
	}

	lineHeight := s.UI().LineHeight()

	// Left side position
	startX := 50.0
	startY := 100.0

	// Player stats
	lines := []string{"Player Info", "-----------"}

	if mover := s.player.GetMover(); mover != nil {
		pos := mover.GetPosition()
		lines = append(lines, fmt.Sprintf("Position: %.0f, %.0f", pos.X, pos.Y))
	}

	if stats := s.player.GetStats(); stats != nil {
		lines = append(lines, fmt.Sprintf("HP: %d / %d", stats.HP, stats.MaxHP))
	} else {
		lines = append(lines, fmt.Sprintf("HP: %d / %d", config.Global.Battle.PlayerHP, config.Global.Battle.PlayerMaxHP))
	}

	for i, line := range lines {
		s.UI().Text(startX, startY+float64(i)*lineHeight, line)
	}

	return nil
}

// renderMenu renders the menu on the right side
func (s *PlayerMenuScene) renderMenu() error {
	menu := s.menuSystem.playerMenu
	if menu == nil {
		return nil
	}

	lineHeight := s.UI().LineHeight()

	// Right side position
	startX := s.GetScreenWidth() - 250.0
	startY := 100.0

	for i, option := range menu.options {
		displayText := "  " + option
		if i == menu.selectedIndex {
			displayText = "> " + option
		}
		s.UI().TextColored(startX, startY+float64(i)*lineHeight, types.Yellow, displayText)
	}

	return nil
}

// GetRenderables returns all game objects in the correct render order
func (s *PlayerMenuScene) GetRenderables() []types.GameObject {
	// Player menu scene doesn't render game objects - just overlays
	return []types.GameObject{}
}

// Cleanup releases scene resources (overrides BaseScene.Cleanup)
func (s *PlayerMenuScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.GetName())

	// Clear menu-specific state
	s.menuSystem = nil
	s.player = nil

	// Call base cleanup (clears layers)
	s.BaseScene.Cleanup()
}

// GetName is inherited from BaseScene

// PlayerMenuSystem manages the player menu UI
type PlayerMenuSystem struct {
	screenWidth  float64
	screenHeight float64
	playerMenu   *PlayerMenu
}

// PlayerMenu represents the player menu with options
type PlayerMenu struct {
	options       []string
	selectedIndex int
}

// NewPlayerMenuSystem creates a new player menu system
func NewPlayerMenuSystem(screenWidth, screenHeight float64) *PlayerMenuSystem {
	return &PlayerMenuSystem{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

// Initialize sets up the player menu system
func (pms *PlayerMenuSystem) Initialize() {
	logger.Logger.Debugf("Initializing player menu system")

	pms.playerMenu = &PlayerMenu{
		options: []string{
			"Save Game",
		},
		selectedIndex: 0,
	}

	logger.Logger.Debugf("Player menu system initialized")
}

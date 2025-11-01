//go:build js

package scenes

import (
	"fmt"
	"syscall/js"

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

// PlayerMenuScene represents the player menu scene accessible from gameplay
type PlayerMenuScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// State change callback (injected by engine)
	stateChangeCallback func(state types.GameState) error

	// Game state manager (injected by engine)
	gameStateManager *gamestate.GameStateManager

	// Player reference (passed from gameplay scene)
	player *gameobject.Player

	// Menu system
	menuSystem *PlayerMenuSystem

	// Game objects organized by layer
	layers map[pkscene.SceneLayer][]types.GameObject

	// Text rendering
	menuFont         text.Font
	menuTextRenderer text.TextRenderer
	canvasManager    canvas.CanvasManager

	// Debug rendering
	debugFont         text.Font
	debugTextRenderer text.TextRenderer

	// Input state tracking
	upPressedLastFrame    bool
	downPressedLastFrame  bool
	enterPressedLastFrame bool
	mPressedLastFrame     bool // M key to close menu
}

// NewPlayerMenuScene creates a new player menu scene
func NewPlayerMenuScene(screenWidth, screenHeight float64) *PlayerMenuScene {
	return &PlayerMenuScene{
		name:         "PlayerMenu",
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		layers:       make(map[pkscene.SceneLayer][]types.GameObject),
	}
}

// SetInputCapturer implements types.SceneInputProvider
func (s *PlayerMenuScene) SetInputCapturer(inputCapturer types.InputCapturer) {
	s.inputCapturer = inputCapturer
}

// SetStateChangeCallback implements types.SceneStateChangeRequester
func (s *PlayerMenuScene) SetStateChangeCallback(callback func(state types.GameState) error) {
	s.stateChangeCallback = callback
}

// SetGameState implements types.SceneGameStateUser
func (s *PlayerMenuScene) SetGameState(gameState interface{}) {
	// Cast to the game's state manager type
	if manager, ok := gameState.(*gamestate.GameStateManager); ok {
		s.gameStateManager = manager
		logger.Logger.Debugf("Player menu scene received game state manager")
	}
}

// SetCanvasManager sets the canvas manager for rendering
func (s *PlayerMenuScene) SetCanvasManager(cm canvas.CanvasManager) {
	s.canvasManager = cm
}

// updatePlayerReference updates the player reference from the game state manager
// Only updates if the player reference has changed to avoid excessive logging
func (s *PlayerMenuScene) updatePlayerReference() {
	if s.gameStateManager == nil {
		if s.player != nil {
			s.player = nil
		}
		return
	}

	// Get player from game state manager
	player := s.gameStateManager.GetPlayer()
	if player == nil {
		if s.player != nil {
			s.player = nil
		}
		return
	}

	// Cast to the game's player type
	if p, ok := player.(*gameobject.Player); ok {
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

// InitializeDebugConsole initializes the debug console font and text renderer
func (s *PlayerMenuScene) InitializeDebugConsole() error {
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

	// Create text renderer
	s.debugTextRenderer = text.NewTextRenderer(s.canvasManager)

	logger.Logger.Debugf("Debug console initialized successfully")
	debug.Console.PostMessage("System", "Player menu ready")

	return nil
}

// InitializeMenuText initializes the menu text rendering system
func (s *PlayerMenuScene) InitializeMenuText() error {
	logger.Logger.Debugf("Initializing menu text rendering for %s scene", s.name)

	// Create and load font metadata for menu text
	s.menuFont = text.NewSpriteFont()
	err := s.menuFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load menu font: %s", err)
		return err
	}

	// Create text renderer for menu
	s.menuTextRenderer = text.NewTextRenderer(s.canvasManager)

	logger.Logger.Debugf("Menu text rendering initialized successfully")
	return nil
}

// GetRequiredAssets implements types.SceneAssetProvider
func (s *PlayerMenuScene) GetRequiredAssets() types.SceneAssets {
	// Font texture path is basePath + ".sheet.png"
	fontTexturePath := config.Global.Debug.FontPath + ".sheet.png"
	return types.SceneAssets{
		TexturePaths: []string{
			fontTexturePath, // Font texture needed for menu text rendering
		},
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	}
}

// Initialize sets up the player menu scene
func (s *PlayerMenuScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.name)

	// Initialize layer slices
	s.layers[pkscene.BACKGROUND] = []types.GameObject{}
	s.layers[pkscene.ENTITIES] = []types.GameObject{}
	s.layers[pkscene.UI] = []types.GameObject{}

	// Initialize menu system
	s.menuSystem = NewPlayerMenuSystem(s.screenWidth, s.screenHeight)
	s.menuSystem.Initialize()

	// Initialize menu text rendering
	err := s.InitializeMenuText()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize menu text: %s", err)
	}

	return nil
}

// Update updates the player menu scene
func (s *PlayerMenuScene) Update(deltaTime float64) {
	// Update player reference from game state (player is part of game state)
	s.updatePlayerReference()

	if s.inputCapturer == nil {
		return
	}

	inputState := s.inputCapturer.GetInputState()

	// Handle menu close (M key)
	if inputState.MPressed && !s.mPressedLastFrame && s.stateChangeCallback != nil {
		logger.Logger.Debugf("M key pressed: closing player menu")
		err := s.stateChangeCallback(types.GAMEPLAY)
		if err != nil {
			logger.Logger.Errorf("Failed to switch back to gameplay: %s", err.Error())
		}
		s.mPressedLastFrame = inputState.MPressed
		return
	}
	s.mPressedLastFrame = inputState.MPressed

	// Handle menu navigation
	menu := s.menuSystem.playerMenu

	// Navigation
	if inputState.UpPressed && !s.upPressedLastFrame {
		menu.selectedIndex--
		if menu.selectedIndex < 0 {
			menu.selectedIndex = len(menu.options) - 1
		}
	}
	if inputState.DownPressed && !s.downPressedLastFrame {
		menu.selectedIndex++
		if menu.selectedIndex >= len(menu.options) {
			menu.selectedIndex = 0
		}
	}
	s.upPressedLastFrame = inputState.UpPressed
	s.downPressedLastFrame = inputState.DownPressed

	// Selection
	if inputState.EnterPressed && !s.enterPressedLastFrame {
		selected := menu.options[menu.selectedIndex]
		if selected == "Save Game" {
			s.handleSaveGame()
		}
	}
	s.enterPressedLastFrame = inputState.EnterPressed

	// Update debug console
	if config.Global.Debug.Enabled {
		debug.Console.Update(deltaTime)
	}
}

// handleSaveGame handles saving the current game state and shows browser alert
func (s *PlayerMenuScene) handleSaveGame() {
	if s.gameStateManager == nil {
		s.showAlert("Save failed: Game state manager not available")
		return
	}

	if s.player == nil {
		s.showAlert("Save failed: Player not available")
		return
	}

	// Get current game state
	currentState := s.gameStateManager.GetState()
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
	s.gameStateManager.UpdateStateFromPlayer(playerPos, playerStats)
	logger.Logger.Debugf("Updated game state before save - Position: (%.2f, %.2f)", playerPos.X, playerPos.Y)

	// Save to localStorage
	saveKey, err := s.gameStateManager.SaveCurrentGame()
	if err != nil {
		logger.Logger.Errorf("Failed to save game: %s", err.Error())
		s.showAlert(fmt.Sprintf("Save failed: %s", err.Error()))
	} else {
		logger.Logger.Infof("Game saved successfully: %s", saveKey)
		s.showAlert("Game saved successfully!")
		debug.Console.PostMessage("System", "Game saved successfully")
	}
}

// showAlert shows a browser alert using syscall/js
func (s *PlayerMenuScene) showAlert(message string) {
	js.Global().Get("window").Call("alert", message)
	logger.Logger.Debugf("Alert shown: %s", message)
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
	if s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	if s.player == nil {
		return nil
	}

	_, cellHeight := s.menuFont.GetCellSize()
	lineHeight := float64(cellHeight)
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		lineHeight *= float64(config.Global.Rendering.PixelScale)
	}
	lineHeight *= config.Global.Rendering.UILineSpacing

	// Left side position
	startX := 50.0
	startY := 100.0

	// Player stats
	var lines []string
	lines = append(lines, "Player Info")
	lines = append(lines, "-----------")

	// Position
	if mover := s.player.GetMover(); mover != nil {
		pos := mover.GetPosition()
		lines = append(lines, fmt.Sprintf("Position: %.0f, %.0f", pos.X, pos.Y))
	}

	// Stats
	if stats := s.player.GetStats(); stats != nil {
		lines = append(lines, fmt.Sprintf("HP: %d / %d", stats.HP, stats.MaxHP))
	} else {
		lines = append(lines, fmt.Sprintf("HP: %d / %d", config.Global.Battle.PlayerHP, config.Global.Battle.PlayerMaxHP))
	}

	// Render each line
	for i, line := range lines {
		err := s.menuTextRenderer.RenderText(
			line,
			types.Vector2{X: startX, Y: startY + float64(i)*lineHeight},
			s.menuFont,
			[4]float32{1.0, 1.0, 1.0, 1.0}, // White text
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render player info line: %s", err)
		}
	}

	return nil
}

// renderMenu renders the menu on the right side
func (s *PlayerMenuScene) renderMenu() error {
	if s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	menu := s.menuSystem.playerMenu
	if menu == nil {
		return nil
	}

	_, cellHeight := s.menuFont.GetCellSize()
	lineHeight := float64(cellHeight)
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		lineHeight *= float64(config.Global.Rendering.PixelScale)
	}
	lineHeight *= config.Global.Rendering.UILineSpacing

	// Right side position
	startX := s.screenWidth - 250.0
	startY := 100.0

	for i, option := range menu.options {
		// Add selection indicator
		displayText := option
		if i == menu.selectedIndex {
			displayText = "> " + option
		} else {
			displayText = "  " + option
		}

		err := s.menuTextRenderer.RenderText(
			displayText,
			types.Vector2{X: startX, Y: startY + float64(i)*lineHeight},
			s.menuFont,
			[4]float32{1.0, 1.0, 0.0, 1.0}, // Yellow text for menu
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render menu item: %s", err)
		}
	}

	return nil
}

// GetRenderables returns all game objects in the correct render order
func (s *PlayerMenuScene) GetRenderables() []types.GameObject {
	// Player menu scene doesn't render game objects - just overlays
	return []types.GameObject{}
}

// Cleanup releases scene resources
func (s *PlayerMenuScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.name)

	// Clear all layers
	for layer := range s.layers {
		s.layers[layer] = nil
	}
	s.layers = make(map[pkscene.SceneLayer][]types.GameObject)
	s.menuSystem = nil
}

// GetName returns the scene identifier
func (s *PlayerMenuScene) GetName() string {
	return s.name
}

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

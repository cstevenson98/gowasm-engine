//go:build js

package scenes

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MenuScene represents the main menu scene with New Game and Load Game options
type MenuScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// State change callback (injected by engine)
	stateChangeCallback func(state types.PipelineState) error

	// Game state manager (injected by engine)
	gameStateManager *gamestate.GameStateManager

	// Menu system
	menuSystem *MainMenuSystem

	// Game objects organized by layer
	layers map[pkscene.SceneLayer][]types.GameObject

	// Text rendering
	menuFont         text.Font
	menuTextRenderer text.TextRenderer
	canvasManager    canvas.CanvasManager

	// Debug rendering
	debugFont         text.Font
	debugTextRenderer text.TextRenderer

	// Debug console toggle state
	f2PressedLastFrame bool

	// Input state tracking
	upPressedLastFrame    bool
	downPressedLastFrame  bool
	enterPressedLastFrame bool

	// Menu mode: "main" or "load"
	menuMode string

	// Load game menu state
	loadGameSaves []gamestate.SaveInfo
	loadGameIndex int // Selected save index in load menu
}

// NewMenuScene creates a new menu scene
func NewMenuScene(screenWidth, screenHeight float64) *MenuScene {
	return &MenuScene{
		name:          "Menu",
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		layers:        make(map[pkscene.SceneLayer][]types.GameObject),
		menuMode:      "main",
		loadGameIndex: 0,
	}
}

// SetInputCapturer implements types.SceneInputProvider
func (s *MenuScene) SetInputCapturer(inputCapturer types.InputCapturer) {
	s.inputCapturer = inputCapturer
}

// SetStateChangeCallback implements types.SceneChangeRequester
func (s *MenuScene) SetStateChangeCallback(callback func(state types.PipelineState) error) {
	s.stateChangeCallback = callback
}

// SetGameState implements types.SceneGameStateUser
func (s *MenuScene) SetGameState(gameState interface{}) {
	// Cast to the game's state manager type
	if manager, ok := gameState.(*gamestate.GameStateManager); ok {
		s.gameStateManager = manager
		logger.Logger.Debugf("Menu scene received game state manager")
	}
}

// SetCanvasManager sets the canvas manager for rendering
func (s *MenuScene) SetCanvasManager(cm canvas.CanvasManager) {
	s.canvasManager = cm
}

// InitializeDebugConsole initializes the debug console font and text renderer
func (s *MenuScene) InitializeDebugConsole() error {
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
	debug.Console.PostMessage("System", "Main menu ready")

	return nil
}

// InitializeMenuText initializes the menu text rendering system
func (s *MenuScene) InitializeMenuText() error {
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
func (s *MenuScene) GetRequiredAssets() types.SceneAssets {
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

// Initialize sets up the menu scene
func (s *MenuScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.name)

	// Initialize layer slices
	s.layers[pkscene.BACKGROUND] = []types.GameObject{}
	s.layers[pkscene.ENTITIES] = []types.GameObject{}
	s.layers[pkscene.UI] = []types.GameObject{}

	// Create black background (BACKGROUND layer)
	// We'll render it as a solid color in RenderOverlays instead of using a texture
	// For now, just ensure layers are initialized

	// Initialize menu system
	s.menuSystem = NewMainMenuSystem(s.screenWidth, s.screenHeight)
	s.menuSystem.Initialize()

	// Initialize menu text rendering
	err := s.InitializeMenuText()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize menu text: %s", err)
	}

	return nil
}

// Update updates the menu scene
func (s *MenuScene) Update(deltaTime float64) {
	if s.inputCapturer == nil {
		return
	}

	inputState := s.inputCapturer.GetInputState()

	// Handle debug console toggle (F2)
	if inputState.F2Pressed && !s.f2PressedLastFrame {
		debug.Console.ToggleVisibility()
		logger.Logger.Debugf("Debug console toggled via F2")
	}
	s.f2PressedLastFrame = inputState.F2Pressed

	// Handle menu navigation based on current mode
	if s.menuMode == "main" {
		s.updateMainMenu(inputState)
	} else if s.menuMode == "load" {
		s.updateLoadMenu(inputState)
	}

	// Update debug console
	if config.Global.Debug.Enabled {
		debug.Console.Update(deltaTime)
	}
}

// updateMainMenu handles input for the main menu
func (s *MenuScene) updateMainMenu(inputState types.InputState) {
	menu := s.menuSystem.mainMenu

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
		if selected == "New Game" {
			if s.gameStateManager != nil {
				err := s.gameStateManager.CreateNewGame()
				if err != nil {
					logger.Logger.Errorf("Failed to create new game: %s", err.Error())
				} else {
					logger.Logger.Debugf("Created new game, switching to gameplay")
					if s.stateChangeCallback != nil {
						err := s.stateChangeCallback(types.GAMEPLAY)
						if err != nil {
							logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
						}
						return
					}
				}
			}
		} else if selected == "Load Game" {
			// Load save list
			if s.gameStateManager != nil {
				saves, err := s.gameStateManager.ListSaves()
				if err != nil {
					logger.Logger.Errorf("Failed to list saves: %s", err.Error())
				} else {
					s.loadGameSaves = saves
					s.loadGameIndex = 0
					s.menuMode = "load"
					logger.Logger.Debugf("Entered load game menu with %d saves", len(saves))
				}
			}
		}
	}
	s.enterPressedLastFrame = inputState.EnterPressed
}

// updateLoadMenu handles input for the load game menu
func (s *MenuScene) updateLoadMenu(inputState types.InputState) {
	// Navigation
	if inputState.UpPressed && !s.upPressedLastFrame {
		s.loadGameIndex--
		if s.loadGameIndex < 0 {
			s.loadGameIndex = len(s.loadGameSaves) - 1
		}
	}
	if inputState.DownPressed && !s.downPressedLastFrame {
		s.loadGameIndex++
		if s.loadGameIndex >= len(s.loadGameSaves) {
			s.loadGameIndex = 0
		}
	}
	s.upPressedLastFrame = inputState.UpPressed
	s.downPressedLastFrame = inputState.DownPressed

	// Selection or back
	if inputState.EnterPressed && !s.enterPressedLastFrame {
		if s.loadGameIndex < len(s.loadGameSaves) {
			// Load selected save
			save := s.loadGameSaves[s.loadGameIndex]
			if s.gameStateManager != nil {
				err := s.gameStateManager.LoadSave(save.Key)
				if err != nil {
					logger.Logger.Errorf("Failed to load save: %s", err.Error())
				} else {
					logger.Logger.Debugf("Loaded save: %s, switching to gameplay", save.Key)
					if s.stateChangeCallback != nil {
						err := s.stateChangeCallback(types.GAMEPLAY)
						if err != nil {
							logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
						}
						return
					}
				}
			}
		}
	}

	// Back to main menu (Escape or special key - for now, if no saves, Enter goes back)
	if len(s.loadGameSaves) == 0 {
		if inputState.EnterPressed && !s.enterPressedLastFrame {
			s.menuMode = "main"
			s.enterPressedLastFrame = inputState.EnterPressed
		}
	}
	s.enterPressedLastFrame = inputState.EnterPressed
}

// RenderOverlays implements types.SceneOverlayRenderer
func (s *MenuScene) RenderOverlays() error {
	// Render black background (solid color - we'll use a simple filled rect approach)
	// For now, just render the menu text

	if s.menuMode == "main" {
		return s.renderMainMenu()
	} else if s.menuMode == "load" {
		return s.renderLoadMenu()
	}

	return nil
}

// renderMainMenu renders the main menu
func (s *MenuScene) renderMainMenu() error {
	if s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	menu := s.menuSystem.mainMenu
	if menu == nil {
		return nil
	}

	// Calculate centered position
	_, cellHeight := s.menuFont.GetCellSize()
	lineHeight := float64(cellHeight)
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		lineHeight *= float64(config.Global.Rendering.PixelScale)
	}
	lineHeight *= config.Global.Rendering.UILineSpacing

	totalHeight := float64(len(menu.options)) * lineHeight
	startY := (s.screenHeight - totalHeight) / 2
	centerX := s.screenWidth / 2

	for i, option := range menu.options {
		// Add selection indicator
		displayText := option
		if i == menu.selectedIndex {
			displayText = "> " + option
		} else {
			displayText = "  " + option
		}

		// Calculate text width for centering (approximate - could be improved)
		textWidth := float64(len(displayText)) * float64(cellHeight) * 0.6 // Rough estimate
		x := centerX - textWidth/2

		err := s.menuTextRenderer.RenderText(
			displayText,
			types.Vector2{X: x, Y: startY + float64(i)*lineHeight},
			s.menuFont,
			[4]float32{1.0, 1.0, 1.0, 1.0}, // White text
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render menu item: %s", err)
		}
	}

	return nil
}

// renderLoadMenu renders the load game menu
func (s *MenuScene) renderLoadMenu() error {
	if s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	// Calculate centered position
	_, cellHeight := s.menuFont.GetCellSize()
	lineHeight := float64(cellHeight)
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		lineHeight *= float64(config.Global.Rendering.PixelScale)
	}
	lineHeight *= config.Global.Rendering.UILineSpacing

	// Title
	title := "Load Game"
	if len(s.loadGameSaves) == 0 {
		title = "No Saves Available"
	}

	totalHeight := float64(len(s.loadGameSaves)+2) * lineHeight // +2 for title and spacing
	startY := (s.screenHeight - totalHeight) / 2
	centerX := s.screenWidth / 2

	// Render title
	titleWidth := float64(len(title)) * float64(cellHeight) * 0.6
	titleX := centerX - titleWidth/2
	err := s.menuTextRenderer.RenderText(
		title,
		types.Vector2{X: titleX, Y: startY},
		s.menuFont,
		[4]float32{1.0, 1.0, 0.0, 1.0}, // Yellow title
	)
	if err != nil {
		logger.Logger.Tracef("Failed to render title: %s", err)
	}

	// Render saves
	for i, save := range s.loadGameSaves {
		displayText := fmt.Sprintf("  %s - Level %d, %d/%d HP", save.DisplayTime, save.PlayerLevel, save.PlayerHP, save.PlayerMaxHP)
		if i == s.loadGameIndex {
			displayText = "> " + displayText[2:] // Replace leading spaces with selection indicator
		}

		textWidth := float64(len(displayText)) * float64(cellHeight) * 0.6
		x := centerX - textWidth/2

		err := s.menuTextRenderer.RenderText(
			displayText,
			types.Vector2{X: x, Y: startY + float64(i+2)*lineHeight}, // +2 for title spacing
			s.menuFont,
			[4]float32{1.0, 1.0, 1.0, 1.0}, // White text
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render save item: %s", err)
		}
	}

	// Render "No saves" message if empty
	if len(s.loadGameSaves) == 0 {
		msg := "Press Enter to return"
		msgWidth := float64(len(msg)) * float64(cellHeight) * 0.6
		msgX := centerX - msgWidth/2
		err := s.menuTextRenderer.RenderText(
			msg,
			types.Vector2{X: msgX, Y: startY + 2*lineHeight},
			s.menuFont,
			[4]float32{0.7, 0.7, 0.7, 1.0}, // Gray text
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render message: %s", err)
		}
	}

	return nil
}

// GetRenderables returns all game objects in the correct render order
func (s *MenuScene) GetRenderables() []types.GameObject {
	// Menu scene doesn't render game objects - just overlays
	return []types.GameObject{}
}

// Cleanup releases scene resources
func (s *MenuScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.name)

	// Clear all layers
	for layer := range s.layers {
		s.layers[layer] = nil
	}
	s.layers = make(map[pkscene.SceneLayer][]types.GameObject)
	s.menuSystem = nil
	s.menuMode = "main"
	s.loadGameSaves = nil
}

// GetName returns the scene identifier
func (s *MenuScene) GetName() string {
	return s.name
}

// MainMenuSystem manages the main menu UI
type MainMenuSystem struct {
	screenWidth  float64
	screenHeight float64
	mainMenu     *MainMenu
}

// MainMenu represents the main menu with options
type MainMenu struct {
	options       []string
	selectedIndex int
}

// NewMainMenuSystem creates a new main menu system
func NewMainMenuSystem(screenWidth, screenHeight float64) *MainMenuSystem {
	return &MainMenuSystem{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

// Initialize sets up the main menu system
func (mms *MainMenuSystem) Initialize() {
	logger.Logger.Debugf("Initializing main menu system")

	mms.mainMenu = &MainMenu{
		options: []string{
			"New Game",
			"Load Game",
		},
		selectedIndex: 0,
	}

	logger.Logger.Debugf("Main menu system initialized")
}

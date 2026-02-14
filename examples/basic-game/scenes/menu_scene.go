//go:build js

package scenes

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MenuScene represents the main menu scene with New Game and Load Game options.
// It embeds BaseScene to inherit all common scene functionality.
type MenuScene struct {
	*pkscene.BaseScene

	// Menu-specific fields
	menuSystem *MainMenuSystem

	// Text rendering
	menuFont         text.Font
	menuTextRenderer text.TextRenderer

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
	baseScene := pkscene.NewBaseScene("Menu", screenWidth, screenHeight)

	// Set required assets
	fontTexturePath := config.Global.Debug.FontPath + ".sheet.png"
	baseScene.SetRequiredAssets(types.SceneAssets{
		TexturePaths: []string{
			fontTexturePath, // Font texture needed for menu text rendering
		},
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	})

	return &MenuScene{
		BaseScene:     baseScene,
		menuMode:      "main",
		loadGameIndex: 0,
	}
}

// All interface implementations (SetInputCapturer, SetStateChangeCallback, SetGameState, SetCanvasManager)
// are inherited from BaseScene

// InitializeMenuText initializes the menu text rendering system
func (s *MenuScene) InitializeMenuText() error {
	logger.Logger.Debugf("Initializing menu text rendering for %s scene", s.GetName())

	// Create and load font metadata for menu text
	s.menuFont = text.NewSpriteFont()
	err := s.menuFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load menu font: %s", err)
		return err
	}

	// Create text renderer for menu using inherited canvasManager
	s.menuTextRenderer = text.NewTextRenderer(s.GetCanvasManager())

	logger.Logger.Debugf("Menu text rendering initialized successfully")
	return nil
}

// GetRequiredAssets is inherited from BaseScene (set in constructor)

// Initialize sets up the menu scene (overrides BaseScene.Initialize)
func (s *MenuScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.GetName())

	// Call base initialization (sets up layers)
	if err := s.BaseScene.Initialize(); err != nil {
		return err
	}

	// Initialize menu system
	s.menuSystem = NewMainMenuSystem(s.GetScreenWidth(), s.GetScreenHeight())
	s.menuSystem.Initialize()

	// Initialize menu text rendering
	if err := s.InitializeMenuText(); err != nil {
		logger.Logger.Warnf("Failed to initialize menu text: %s", err)
	}

	return nil
}

// Update updates the menu scene (overrides BaseScene.Update)
func (s *MenuScene) Update(deltaTime float64) {
	// Get input state using inherited method
	inputState := s.GetInputState()

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
			gameState := s.GetGameState()
			if gameState != nil {
				if manager, ok := gameState.(*gamestate.GameStateManager); ok {
					err := manager.CreateNewGame()
					if err != nil {
						logger.Logger.Errorf("Failed to create new game: %s", err.Error())
					} else {
						logger.Logger.Debugf("Created new game, switching to gameplay")
						err := s.RequestStateChange(types.GAMEPLAY)
						if err != nil {
							logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
						}
						return
					}
				}
			}
		} else if selected == "Load Game" {
			// Load save list
			gameState := s.GetGameState()
			if gameState != nil {
				if manager, ok := gameState.(*gamestate.GameStateManager); ok {
					saves, err := manager.ListSaves()
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
			gameState := s.GetGameState()
			if gameState != nil {
				if manager, ok := gameState.(*gamestate.GameStateManager); ok {
					err := manager.LoadSave(save.Key)
					if err != nil {
						logger.Logger.Errorf("Failed to load save: %s", err.Error())
					} else {
						logger.Logger.Debugf("Loaded save: %s, switching to gameplay", save.Key)
						err := s.RequestStateChange(types.GAMEPLAY)
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
	// Render menu first
	if s.menuMode == "main" {
		if err := s.renderMainMenu(); err != nil {
			return err
		}
	} else if s.menuMode == "load" {
		if err := s.renderLoadMenu(); err != nil {
			return err
		}
	}

	// Then render debug console (inherited from BaseScene)
	return s.BaseScene.RenderOverlays()
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
	startY := (s.GetScreenHeight() - totalHeight) / 2
	centerX := s.GetScreenWidth() / 2

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
	startY := (s.GetScreenHeight() - totalHeight) / 2
	centerX := s.GetScreenWidth() / 2

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
// GetRenderables is inherited from BaseScene (menu doesn't render game objects, just overlays)

// Cleanup releases scene resources (overrides BaseScene.Cleanup)
func (s *MenuScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.GetName())

	// Clear menu-specific state
	s.menuSystem = nil
	s.menuMode = "main"
	s.loadGameSaves = nil

	// Call base cleanup (clears layers)
	s.BaseScene.Cleanup()
}

// GetName is inherited from BaseScene

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

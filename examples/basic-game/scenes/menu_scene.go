package scenes

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MenuScene represents the main menu scene with New Game and Load Game options.
// It embeds BaseScene to inherit all common scene functionality.
type MenuScene struct {
	*pkscene.BaseScene

	// Menu-specific fields
	menuSystem *MainMenuSystem

	// Menu mode: "main" or "load"
	menuMode string

	// Load game menu state
	loadGameSaves []gamestate.SaveInfo
	loadGameIndex int // Selected save index in load menu
}

// NewMenuScene creates a new menu scene
func NewMenuScene(screenWidth, screenHeight float64) *MenuScene {
	baseScene := pkscene.NewBaseScene("Menu", screenWidth, screenHeight)

	// Declare required assets. The font texture is loaded automatically from
	// FontPaths, so it doesn't need to be listed under TexturePaths.
	baseScene.SetRequiredAssets(types.SceneAssets{
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

// All interface implementations (SetInputCapturer, SetStateChangeCallback, SetGameState)
// are inherited from BaseScene

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

	return nil
}

// Update updates the menu scene (overrides BaseScene.Update)
func (s *MenuScene) Update(deltaTime float64) {
	// Get input state using inherited method
	inputState := s.GetInputState()

	// Handle debug console toggle (F2)
	if inputState.F2Pressed && !inputState.F2PressedLastFrame {
		debug.Console.ToggleVisibility()
		logger.Logger.Debugf("Debug console toggled via F2")
	}

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
}

// updateLoadMenu handles input for the load game menu
func (s *MenuScene) updateLoadMenu(inputState types.InputState) {
	// Navigation
	if inputState.UpPressed && !inputState.UpPressedLastFrame {
		s.loadGameIndex--
		if s.loadGameIndex < 0 {
			s.loadGameIndex = len(s.loadGameSaves) - 1
		}
	}
	if inputState.DownPressed && !inputState.DownPressedLastFrame {
		s.loadGameIndex++
		if s.loadGameIndex >= len(s.loadGameSaves) {
			s.loadGameIndex = 0
		}
	}

	// Selection or back
	if inputState.EnterPressed && !inputState.EnterPressedLastFrame {
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

	// Back to main menu (for now, if no saves, Enter goes back)
	if len(s.loadGameSaves) == 0 {
		if inputState.EnterPressed && !inputState.EnterPressedLastFrame {
			s.menuMode = "main"
		}
	}
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
	menu := s.menuSystem.mainMenu
	if menu == nil {
		return nil
	}

	lineHeight := s.UI().LineHeight()
	totalHeight := float64(len(menu.options)) * lineHeight
	startY := (s.GetScreenHeight() - totalHeight) / 2

	for i, option := range menu.options {
		displayText := "  " + option
		if i == menu.selectedIndex {
			displayText = "> " + option
		}
		s.UI().TextCentered(startY+float64(i)*lineHeight, types.White, displayText)
	}

	return nil
}

// renderLoadMenu renders the load game menu
func (s *MenuScene) renderLoadMenu() error {
	lineHeight := s.UI().LineHeight()

	title := "Load Game"
	if len(s.loadGameSaves) == 0 {
		title = "No Saves Available"
	}

	totalHeight := float64(len(s.loadGameSaves)+2) * lineHeight // +2 for title and spacing
	startY := (s.GetScreenHeight() - totalHeight) / 2

	s.UI().TextCentered(startY, types.Yellow, title)

	for i, save := range s.loadGameSaves {
		displayText := fmt.Sprintf("  %s - Level %d, %d/%d HP", save.DisplayTime, save.PlayerLevel, save.PlayerHP, save.PlayerMaxHP)
		if i == s.loadGameIndex {
			displayText = "> " + displayText[2:] // Replace leading spaces with selection indicator
		}
		s.UI().TextCentered(startY+float64(i+2)*lineHeight, types.White, displayText)
	}

	if len(s.loadGameSaves) == 0 {
		s.UI().TextCentered(startY+2*lineHeight, types.Gray, "Press Enter to return")
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

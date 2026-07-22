// Package states holds the game's top-level states (menu, gameplay, player menu,
// battle). Each embeds state.BaseState, so it owns an ecs.World and a system
// schedule and gets dependency accessors (Input, UI, RequestState, ...) for free.
package states

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// MenuState is the main menu with New Game and Load Game options.
type MenuState struct {
	*state.BaseState

	options       []string
	selectedIndex int

	// Load submenu.
	menuMode      string // "main" or "load"
	loadGameSaves []gamestate.SaveInfo
	loadGameIndex int
}

// NewMenuState creates the menu state.
func NewMenuState() *MenuState {
	return &MenuState{
		BaseState: state.NewBaseState("Menu"),
		menuMode:  "main",
	}
}

// Enter initialises the menu.
func (s *MenuState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}
	s.options = []string{"New Game", "Load Game"}
	s.selectedIndex = 0
	s.menuMode = "main"
	s.loadGameSaves = nil
	s.loadGameIndex = 0
	logger.Logger.Debugf("Initialized %s state", s.Name())
	return nil
}

// Update handles menu navigation and selection.
func (s *MenuState) Update(dt float64) {
	in := s.Input()
	if s.menuMode == "main" {
		s.updateMainMenu(in)
	} else {
		s.updateLoadMenu(in)
	}
	s.BaseState.Update(dt)
}

func (s *MenuState) manager() *gamestate.GameStateManager {
	if m, ok := s.GameStateProvider().(*gamestate.GameStateManager); ok {
		return m
	}
	return nil
}

func (s *MenuState) updateMainMenu(in types.InputState) {
	if in.UpPressed && !in.UpPressedLastFrame {
		s.selectedIndex = (s.selectedIndex - 1 + len(s.options)) % len(s.options)
	}
	if in.DownPressed && !in.DownPressedLastFrame {
		s.selectedIndex = (s.selectedIndex + 1) % len(s.options)
	}
	if in.EnterPressed && !in.EnterPressedLastFrame {
		switch s.options[s.selectedIndex] {
		case "New Game":
			if m := s.manager(); m != nil {
				if err := m.CreateNewGame(); err != nil {
					logger.Logger.Errorf("Failed to create new game: %s", err.Error())
					return
				}
				if err := s.RequestState(types.GAMEPLAY); err != nil {
					logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
				}
			}
		case "Load Game":
			if m := s.manager(); m != nil {
				saves, err := m.ListSaves()
				if err != nil {
					logger.Logger.Errorf("Failed to list saves: %s", err.Error())
					return
				}
				s.loadGameSaves = saves
				s.loadGameIndex = 0
				s.menuMode = "load"
			}
		}
	}
}

func (s *MenuState) updateLoadMenu(in types.InputState) {
	if len(s.loadGameSaves) > 0 {
		if in.UpPressed && !in.UpPressedLastFrame {
			s.loadGameIndex = (s.loadGameIndex - 1 + len(s.loadGameSaves)) % len(s.loadGameSaves)
		}
		if in.DownPressed && !in.DownPressedLastFrame {
			s.loadGameIndex = (s.loadGameIndex + 1) % len(s.loadGameSaves)
		}
	}

	if in.EnterPressed && !in.EnterPressedLastFrame {
		if len(s.loadGameSaves) == 0 {
			s.menuMode = "main"
			return
		}
		save := s.loadGameSaves[s.loadGameIndex]
		if m := s.manager(); m != nil {
			if err := m.LoadSave(save.Key); err != nil {
				logger.Logger.Errorf("Failed to load save: %s", err.Error())
				return
			}
			if err := s.RequestState(types.GAMEPLAY); err != nil {
				logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
			}
		}
	}
}

// DrawOverlays renders the menu and the debug console.
func (s *MenuState) DrawOverlays() error {
	if s.menuMode == "main" {
		s.renderMainMenu()
	} else {
		s.renderLoadMenu()
	}
	return s.BaseState.DrawOverlays()
}

func (s *MenuState) renderMainMenu() {
	ui := s.UI()
	lineHeight := ui.LineHeight()
	totalHeight := float64(len(s.options)) * lineHeight
	startY := (s.ScreenHeight() - totalHeight) / 2
	for i, option := range s.options {
		text := "  " + option
		if i == s.selectedIndex {
			text = "> " + option
		}
		ui.TextCentered(startY+float64(i)*lineHeight, types.White, text)
	}
}

func (s *MenuState) renderLoadMenu() {
	ui := s.UI()
	lineHeight := ui.LineHeight()

	title := "Load Game"
	if len(s.loadGameSaves) == 0 {
		title = "No Saves Available"
	}
	totalHeight := float64(len(s.loadGameSaves)+2) * lineHeight
	startY := (s.ScreenHeight() - totalHeight) / 2
	ui.TextCentered(startY, types.Yellow, title)

	for i, save := range s.loadGameSaves {
		text := fmt.Sprintf("  %s - Level %d, %d/%d HP", save.DisplayTime, save.PlayerLevel, save.PlayerHP, save.PlayerMaxHP)
		if i == s.loadGameIndex {
			text = "> " + text[2:]
		}
		ui.TextCentered(startY+float64(i+2)*lineHeight, types.White, text)
	}

	if len(s.loadGameSaves) == 0 {
		ui.TextCentered(startY+2*lineHeight, types.Gray, "Press Enter to return")
	}
}

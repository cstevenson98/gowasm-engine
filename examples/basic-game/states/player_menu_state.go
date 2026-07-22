package states

import (
	"fmt"

	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// PlayerMenuState is the in-game menu (opened with M) showing player info and a
// save option. It has no world entities; player data comes from the persistent
// game state, which gameplay keeps up to date.
type PlayerMenuState struct {
	*state.BaseState

	options       []string
	selectedIndex int
}

// NewPlayerMenuState creates the player menu state.
func NewPlayerMenuState() *PlayerMenuState {
	return &PlayerMenuState{BaseState: state.NewBaseState("PlayerMenu")}
}

// GetRequiredAssets declares the font used for menu text.
func (s *PlayerMenuState) GetRequiredAssets() state.Assets {
	return state.Assets{FontPaths: []string{config.Global.Debug.FontPath}}
}

// Enter initialises the menu options.
func (s *PlayerMenuState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}
	s.options = []string{"Save Game"}
	s.selectedIndex = 0
	return nil
}

func (s *PlayerMenuState) manager() *gamestate.GameStateManager {
	if m, ok := s.GameStateProvider().(*gamestate.GameStateManager); ok {
		return m
	}
	return nil
}

// Update handles navigation, selection, and closing the menu.
func (s *PlayerMenuState) Update(dt float64) {
	in := s.Input()

	if in.MPressed && !in.MPressedLastFrame {
		if err := s.RequestState(types.GAMEPLAY); err != nil {
			logger.Logger.Errorf("Failed to return to gameplay: %s", err.Error())
		}
		return
	}

	if in.UpPressed && !in.UpPressedLastFrame {
		s.selectedIndex = (s.selectedIndex - 1 + len(s.options)) % len(s.options)
	}
	if in.DownPressed && !in.DownPressedLastFrame {
		s.selectedIndex = (s.selectedIndex + 1) % len(s.options)
	}
	if in.EnterPressed && !in.EnterPressedLastFrame {
		if s.options[s.selectedIndex] == "Save Game" {
			s.handleSaveGame()
		}
	}

	s.BaseState.Update(dt)
}

func (s *PlayerMenuState) handleSaveGame() {
	m := s.manager()
	if m == nil {
		s.showAlert("Save failed: game state manager not available")
		return
	}
	if m.GetState() == nil {
		s.showAlert("Save failed: no game state (create a new game first)")
		return
	}
	saveKey, err := m.SaveCurrentGame()
	if err != nil {
		logger.Logger.Errorf("Failed to save game: %s", err.Error())
		s.showAlert(fmt.Sprintf("Save failed: %s", err.Error()))
		return
	}
	logger.Logger.Infof("Game saved: %s", saveKey)
	s.showAlert("Game saved successfully!")
}

func (s *PlayerMenuState) showAlert(message string) {
	logger.Logger.Infof("Alert: %s", message)
	debug.Console.PostMessage("Alert", message)
}

// DrawOverlays renders player info, the menu, and the debug console.
func (s *PlayerMenuState) DrawOverlays() error {
	s.renderPlayerInfo()
	s.renderMenu()
	return s.BaseState.DrawOverlays()
}

func (s *PlayerMenuState) renderPlayerInfo() {
	m := s.manager()
	if m == nil || m.GetState() == nil {
		return
	}
	gs := m.GetState()
	ui := s.UI()
	lineHeight := ui.LineHeight()
	startX, startY := 50.0, 100.0

	lines := []string{
		"Player Info",
		"-----------",
		fmt.Sprintf("Position: %.0f, %.0f", gs.PlayerPosition.X, gs.PlayerPosition.Y),
		fmt.Sprintf("HP: %d / %d", gs.PlayerStats.HP, gs.PlayerStats.MaxHP),
		fmt.Sprintf("Level: %d", gs.PlayerStats.Level),
	}
	for i, line := range lines {
		ui.Text(startX, startY+float64(i)*lineHeight, line)
	}
}

func (s *PlayerMenuState) renderMenu() {
	ui := s.UI()
	lineHeight := ui.LineHeight()
	startX := s.ScreenWidth() - 250.0
	startY := 100.0
	for i, option := range s.options {
		text := "  " + option
		if i == s.selectedIndex {
			text = "> " + option
		}
		ui.TextColored(startX, startY+float64(i)*lineHeight, types.Yellow, text)
	}
}

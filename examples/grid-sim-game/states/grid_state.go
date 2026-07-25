package states

import (
	"example.com/grid-sim-game/game/entities"
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GridState is the (only) state of the grid-sim-game example: a scrollable
// grid the player populates with generators, houses and lines via a
// clickable toolbar.
type GridState struct {
	*state.BaseState
}

// NewGridState creates the grid state.
func NewGridState() *GridState {
	return &GridState{BaseState: state.NewBaseState("Grid")}
}

// Enter builds the grid (one blank tile per cell) and registers the
// placement and camera-scroll systems.
func (s *GridState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	ecs.SetResource(s.World(), &entities.PlacementState{Tool: entities.ToolNone})
	ecs.SetResource(s.World(), entities.NewGridOccupancy())

	cfg := gameconfig.Global
	for row := 0; row < cfg.GridRows; row++ {
		for col := 0; col < cfg.GridCols; col++ {
			entities.SpawnBlank(s.World(), entities.GridCoord{Col: col, Row: row})
		}
	}

	// Placement resolves clicks against the camera position from the frame
	// just rendered, then the camera scrolls for the next frame.
	s.Schedule().
		Add(entities.NewPlacementSystem(s.World())).
		Add(entities.NewCameraScrollSystem(cfg.CameraSpeed))

	return nil
}

// DrawOverlays draws the toolbar (with the selected tool highlighted and a
// hint while a line is pending) plus the debug console.
func (s *GridState) DrawOverlays() error {
	s.renderToolbar()
	return s.BaseState.DrawOverlays()
}

func (s *GridState) renderToolbar() {
	cfg := gameconfig.Global
	ui := s.UI()

	ui.Rect(0, 0, s.ScreenWidth(), cfg.ToolbarHeight, types.Color{0.1, 0.1, 0.12, 1})

	placement := ecs.GetResource[entities.PlacementState](s.World())
	selected := entities.ToolNone
	linePending := false
	if placement != nil {
		selected = placement.Tool
		linePending = placement.LinePending
	}

	for _, b := range entities.ToolbarButtons() {
		bg := types.Color{0.3, 0.3, 0.34, 1}
		fg := types.White
		if b.Tool == selected {
			bg = types.Yellow
			fg = types.Black
		}
		ui.Rect(b.X, b.Y, b.W, b.H, bg)
		ui.TextColored(b.X+3, b.Y+2, fg, b.Label)
	}

	if linePending {
		ui.TextColored(cfg.ButtonMarginX, cfg.ToolbarHeight+4, types.Yellow, "Line: click the end cell")
	}
}

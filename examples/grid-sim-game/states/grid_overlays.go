package states

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

func (s *GridState) renderToolbar() {
	cfg := gameconfig.Global
	ui := s.UI()
	playW := cfg.PlayfieldWidth(s.ScreenWidth())

	ui.Rect(0, 0, playW, cfg.ToolbarHeight, types.Color{0.1, 0.1, 0.12, 1})

	placement := ecs.GetResource[grid.PlacementState](s.World())
	selected := grid.ToolNone
	linePending := false
	if placement != nil {
		selected = placement.Tool
		linePending = placement.LinePending
	}

	for _, b := range grid.ToolbarButtons() {
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
	} else if selected == grid.ToolNone {
		ui.TextColored(cfg.ButtonMarginX, cfg.ToolbarHeight+4, types.Color{0.7, 0.7, 0.75, 1}, "Click cell to inspect · C clears tool")
	}
}

// renderGridChrome draws hover/selection borders and the placement ghost.
func (s *GridState) renderGridChrome() {
	placement := ecs.GetResource[grid.PlacementState](s.World())
	cam := ecs.GetResource[components.Camera](s.World())
	occupancy := ecs.GetResource[grid.GridOccupancy](s.World())
	if placement == nil || cam == nil {
		return
	}
	ui := s.UI()

	if placement.HasSelection {
		s.drawCellBorder(ui, cam, placement.SelectedCell, types.Color{0.2, 0.85, 1, 1}, 2)
	}
	if placement.HoverValid {
		s.drawCellBorder(ui, cam, placement.HoverCell, types.Color{1, 1, 0.35, 1}, 2)
	}
	if placement.Tool != grid.ToolNone && placement.HoverValid {
		s.drawPlacementGhost(ui, cam, occupancy, placement)
	}
}

func (s *GridState) drawCellBorder(
	ui types.UIManager,
	cam *components.Camera,
	cell grid.GridCoord,
	c types.Color,
	thickness float64,
) {
	x, y, w, h := grid.CellScreenRect(cam, cell)
	ui.Rect(x, y, w, thickness, c)
	ui.Rect(x, y+h-thickness, w, thickness, c)
	ui.Rect(x, y, thickness, h, c)
	ui.Rect(x+w-thickness, y, thickness, h, c)
}

func (s *GridState) drawPlacementGhost(
	ui types.UIManager,
	cam *components.Camera,
	occupancy *grid.GridOccupancy,
	placement *grid.PlacementState,
) {
	cells := []grid.GridCoord{placement.HoverCell}
	if placement.Tool == grid.ToolLine && placement.LinePending {
		cells = grid.ManhattanPath(placement.LineStart, placement.HoverCell)
	}

	letter := placement.Tool.GhostLetter()
	for _, cell := range cells {
		occupied := occupancy != nil && occupancy.Occupied(cell)
		fill, fg := ghostColors(placement.Tool, occupied)
		x, y, w, h := grid.CellScreenRect(cam, cell)
		ui.Rect(x, y, w, h, fill)
		ui.TextColored(x+w*0.3, y+h*0.25, fg, letter)
	}
}

func ghostColors(t grid.Tool, occupied bool) (fill, fg types.Color) {
	switch t {
	case grid.ToolDelete:
		if occupied {
			return types.Color{0.7, 0.1, 0.1, 0.45}, types.Color{1, 0.25, 0.25, 1}
		}
		return types.Color{0.3, 0.3, 0.3, 0.35}, types.Color{0.55, 0.55, 0.55, 1}
	default:
		if occupied {
			return types.Color{0.7, 0.15, 0.1, 0.4}, types.Color{1, 0.35, 0.3, 1}
		}
		return types.Color{0.15, 0.55, 0.25, 0.4}, types.Color{0.85, 1, 0.85, 1}
	}
}

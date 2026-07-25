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

// renderGridChrome draws the procedural playfield grid, polyline line fills,
// hover/selection borders, and the placement ghost.
func (s *GridState) renderGridChrome() {
	placement := ecs.GetResource[grid.PlacementState](s.World())
	cam := ecs.GetResource[components.Camera](s.World())
	occupancy := ecs.GetResource[grid.GridOccupancy](s.World())
	if cam == nil {
		return
	}
	ui := s.UI()

	s.renderGridBackground(ui, cam)
	s.renderLinePaths(ui, cam)

	if placement == nil {
		return
	}
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

// renderGridBackground fills the playfield and draws grid lines with
// O(visible cols + rows) rects — not O(cells). Per-cell fills were thousands
// of immediate-mode draws per frame and crushed FPS.
func (s *GridState) renderGridBackground(ui types.UIManager, cam *components.Camera) {
	cfg := gameconfig.Global
	playW := cfg.PlayfieldWidth(s.ScreenWidth())
	screenH := s.ScreenHeight()
	ts := cfg.TileSize
	zoom := cam.Zoom
	if zoom <= 0 {
		zoom = 1
	}

	// No opaque fill here: overlays draw after world sprites and would hide
	// placed entities. Playfield tone comes from a LayerBackground entity.

	worldX0 := cam.X
	worldY0 := cam.Y + cfg.ToolbarHeight/zoom
	worldX1 := cam.X + playW/zoom
	worldY1 := cam.Y + screenH/zoom

	c0 := int(worldX0 / ts)
	r0 := int(worldY0 / ts)
	// Inclusive line indices: draw the far edge of the last visible cell too.
	c1 := int(worldX1/ts) + 1
	r1 := int(worldY1/ts) + 1
	if c0 < 0 {
		c0 = 0
	}
	if r0 < 0 {
		r0 = 0
	}
	if c1 > cfg.GridCols {
		c1 = cfg.GridCols
	}
	if r1 > cfg.GridRows {
		r1 = cfg.GridRows
	}
	if c0 > c1 || r0 > r1 {
		return
	}

	line := types.Color{0.22, 0.22, 0.26, 1}
	// Vertical lines spanning the visible row band.
	y0 := (float64(r0)*ts - cam.Y) * zoom
	y1 := (float64(r1)*ts - cam.Y) * zoom
	if y0 < cfg.ToolbarHeight {
		y0 = cfg.ToolbarHeight
	}
	h := y1 - y0
	if h > 0 {
		for col := c0; col <= c1; col++ {
			x := (float64(col)*ts - cam.X) * zoom
			ui.Rect(x, y0, 1, h, line)
		}
	}
	// Horizontal lines spanning the visible column band.
	x0 := (float64(c0)*ts - cam.X) * zoom
	x1 := (float64(c1)*ts - cam.X) * zoom
	if x0 < 0 {
		x0 = 0
	}
	if x1 > playW {
		x1 = playW
	}
	w := x1 - x0
	if w > 0 {
		for row := r0; row <= r1; row++ {
			y := (float64(row)*ts - cam.Y) * zoom
			if y < cfg.ToolbarHeight {
				continue
			}
			ui.Rect(x0, y, w, 1, line)
		}
	}
}

// renderLinePaths fills secondary cells of polyline lines (path[0] already
// has an ECS sprite).
func (s *GridState) renderLinePaths(ui types.UIManager, cam *components.Camera) {
	if s.linePathFilter == nil {
		s.linePathFilter = ecs.NewFilter1[grid.LinePath](s.World())
	}
	lineFill := types.Color{0.35, 0.55, 0.75, 0.85}
	s.linePathFilter.Each(func(_ ecs.Entity, lp *grid.LinePath) {
		for i, cell := range lp.Cells {
			if i == 0 {
				continue // sprite on path[0]
			}
			x, y, w, h := grid.CellScreenRect(cam, cell)
			ui.Rect(x, y, w, h, lineFill)
		}
	})
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

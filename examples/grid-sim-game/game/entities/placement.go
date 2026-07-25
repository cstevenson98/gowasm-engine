package entities

import (
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// PlacementSystem is the only system that reads mouse clicks and turns them
// into toolbar tool selection or grid placement. It mutates PlacementState
// (the selected tool / pending line start) and GridOccupancy, and spawns the
// placed entities directly.
type PlacementSystem struct{}

// NewPlacementSystem builds the system. It holds no per-World state itself
// (everything it needs lives in resources), so w is unused; it is accepted
// for consistency with the engine's "build systems from a World" convention.
func NewPlacementSystem(_ *ecs.World) *PlacementSystem {
	return &PlacementSystem{}
}

// Update handles at most one click per frame (mouse buttons are edge
// detected via Mouse.Left.Pressed/PressedLastFrame).
func (s *PlacementSystem) Update(w *ecs.World, dt float64) {
	in := ecs.GetResource[components.Input](w)
	placement := ecs.GetResource[PlacementState](w)
	occupancy := ecs.GetResource[GridOccupancy](w)
	cam := ecs.GetResource[components.Camera](w)
	if in == nil || placement == nil || occupancy == nil || cam == nil {
		return
	}

	mouse := in.State.Mouse
	if !mouse.Left.Pressed || mouse.Left.PressedLastFrame {
		return
	}

	if mouse.Y < gameconfig.Global.ToolbarHeight {
		s.handleToolbarClick(placement, mouse.X, mouse.Y)
		return
	}
	s.handleGridClick(w, placement, occupancy, cam, mouse.X, mouse.Y)
}

// handleToolbarClick selects the clicked button's tool, or deselects it
// (ToolNone) if it was already selected. Any pending line is cancelled
// whenever the tool changes.
func (s *PlacementSystem) handleToolbarClick(placement *PlacementState, x, y float64) {
	for _, b := range ToolbarButtons() {
		if !b.Contains(x, y) {
			continue
		}
		if placement.Tool == b.Tool {
			placement.Tool = ToolNone
		} else {
			placement.Tool = b.Tool
		}
		placement.LinePending = false
		return
	}
}

// handleGridClick converts a click below the toolbar into a grid cell (via
// the current camera offset) and dispatches to the placement logic for the
// currently selected tool.
func (s *PlacementSystem) handleGridClick(w *ecs.World, placement *PlacementState, occupancy *GridOccupancy, cam *components.Camera, x, y float64) {
	ts := gameconfig.Global.TileSize
	cell := GridCoord{
		Col: int((x + cam.X) / ts),
		Row: int((y + cam.Y) / ts),
	}
	if cell.Col < 0 || cell.Row < 0 || cell.Col >= gameconfig.Global.GridCols || cell.Row >= gameconfig.Global.GridRows {
		return
	}

	switch placement.Tool {
	case ToolGenerator:
		s.placeSingle(w, occupancy, cell, SpawnGenerator)
	case ToolHouse:
		s.placeSingle(w, occupancy, cell, SpawnHouse)
	case ToolLine:
		s.handleLineClick(w, placement, occupancy, cell)
	}
}

// placeSingle spawns a one-cell entity via spawn if cell is free.
func (s *PlacementSystem) placeSingle(w *ecs.World, occupancy *GridOccupancy, cell GridCoord, spawn func(*ecs.World, GridCoord) ecs.Entity) {
	if occupancy.Occupied(cell) {
		logger.Logger.Debugf("grid-sim: cell %+v is occupied, ignoring placement", cell)
		return
	}
	occupancy.Occupy(cell, spawn(w, cell))
}

// handleLineClick implements the two-click line flow: the first click on a
// free cell records it as the pending start; the second computes the
// Manhattan path to it and, if every cell on that path is free, spawns a
// line-tile entity per cell. An invalid second click (blocked path) simply
// cancels the pending line rather than erroring.
func (s *PlacementSystem) handleLineClick(w *ecs.World, placement *PlacementState, occupancy *GridOccupancy, cell GridCoord) {
	if !placement.LinePending {
		if occupancy.Occupied(cell) {
			return
		}
		placement.LineStart = cell
		placement.LinePending = true
		return
	}

	placement.LinePending = false

	path := ManhattanPath(placement.LineStart, cell)
	for _, c := range path {
		if occupancy.Occupied(c) {
			logger.Logger.Debugf("grid-sim: line path blocked at %+v, cancelling", c)
			return
		}
	}
	for _, c := range path {
		occupancy.Occupy(c, SpawnLineSegment(w, c))
	}
}

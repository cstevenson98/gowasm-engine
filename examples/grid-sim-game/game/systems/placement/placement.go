// Package placement implements PlacementSystem: toolbar tool selection and
// grid spawn/delete/line placement. Hover and ToolNone selection live in
// package pointer; electrical join/leave is delegated to package wiring.
package placement

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/gameconfig"
	"example.com/grid-sim-game/game/systems/wiring"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// PlacementSystem turns toolbar clicks and active-tool grid clicks into
// spawn/delete/line actions. It mutates PlacementState.Tool / Line* and
// GridOccupancy; pointer owns hover/selection.
type PlacementSystem struct{}

// NewPlacementSystem builds the system.
func NewPlacementSystem(_ *ecs.World) *PlacementSystem {
	return &PlacementSystem{}
}

// Update handles toolbar clicks and placement when a tool is active.
// Hover must already be refreshed by PointerSystem earlier in the schedule.
func (s *PlacementSystem) Update(w *ecs.World, _ float64) {
	in := ecs.GetResource[components.Input](w)
	placement := ecs.GetResource[grid.PlacementState](w)
	occupancy := ecs.GetResource[grid.GridOccupancy](w)
	cam := ecs.GetResource[components.Camera](w)
	if in == nil || placement == nil || occupancy == nil || cam == nil {
		return
	}

	mouse := in.State.Mouse
	if !mouse.Left.Pressed || mouse.Left.PressedLastFrame {
		return
	}

	bounds := ecs.GetResource[components.ScreenBounds](w)
	if bounds != nil && mouse.X >= gameconfig.Global.PlayfieldWidth(bounds.W) {
		return
	}

	if mouse.Y < gameconfig.Global.ToolbarHeight {
		handleToolbarClick(placement, mouse.X, mouse.Y)
		return
	}

	if placement.Tool == grid.ToolNone {
		return
	}

	s.handleGridClick(w, placement, occupancy, cam, mouse.X, mouse.Y)
}

func handleToolbarClick(placement *grid.PlacementState, x, y float64) {
	for _, b := range grid.ToolbarButtons() {
		if !b.Contains(x, y) {
			continue
		}
		if placement.Tool == b.Tool {
			placement.Tool = grid.ToolNone
		} else {
			placement.Tool = b.Tool
		}
		placement.LinePending = false
		return
	}
}

func (s *PlacementSystem) handleGridClick(w *ecs.World, placement *grid.PlacementState, occupancy *grid.GridOccupancy, cam *components.Camera, x, y float64) {
	cell, ok := grid.ScreenToCell(cam, x, y)
	if !ok {
		return
	}

	switch placement.Tool {
	case grid.ToolGenerator:
		if e, ok := placeSingle(w, occupancy, cell, grid.SpawnGenerator); ok {
			wiring.Attach(w, e, grid.ToolGenerator, cell, occupancy)
		}
	case grid.ToolHouse:
		if e, ok := placeSingle(w, occupancy, cell, grid.SpawnHouse); ok {
			wiring.Attach(w, e, grid.ToolHouse, cell, occupancy)
		}
	case grid.ToolDelete:
		deleteCell(w, occupancy, cell)
	case grid.ToolLine:
		s.handleLineClick(w, placement, occupancy, cell)
	}
}

func placeSingle(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord, spawn func(*ecs.World, grid.GridCoord) ecs.Entity) (ecs.Entity, bool) {
	if occupancy.Occupied(cell) {
		logger.Logger.Debugf("grid-sim: cell %+v occupied, ignoring place", cell)
		return ecs.Entity{}, false
	}
	e := spawn(w, cell)
	occupancy.Occupy(cell, e)
	return e, true
}

func (s *PlacementSystem) handleLineClick(w *ecs.World, placement *grid.PlacementState, occupancy *grid.GridOccupancy, cell grid.GridCoord) {
	if !placement.LinePending {
		placement.LinePending = true
		placement.LineStart = cell
		return
	}
	placement.LinePending = false
	path := grid.ManhattanPath(placement.LineStart, cell)
	for _, c := range path {
		if occupancy.Occupied(c) {
			return // abort whole stroke if any cell blocked
		}
	}
	e := grid.SpawnLine(w, path)
	if e == (ecs.Entity{}) {
		return
	}
	for _, c := range path {
		occupancy.Occupy(c, e)
	}
	wiring.Attach(w, e, grid.ToolLine, path[0], occupancy)
}

// deleteCell removes the occupant at cell (whole polyline if LinePath),
// detaches from the electrical network, and removes the ECS entity.
func deleteCell(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord) bool {
	e, ok := occupancy.Cells[cell]
	if !ok {
		logger.Logger.Debugf("grid-sim: delete on empty cell %+v, ignoring", cell)
		return false
	}
	if lp := ecs.NewMap1[grid.LinePath](w).Get(e); lp != nil {
		for _, c := range lp.Cells {
			delete(occupancy.Cells, c)
		}
	} else {
		delete(occupancy.Cells, cell)
	}
	wiring.Detach(w, e)
	w.Remove(e)
	return true
}

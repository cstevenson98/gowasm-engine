// Package placement implements PlacementSystem, which translates mouse clicks
// into toolbar tool selection, cell selection, and grid entity spawning.
package placement

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/gameconfig"
	"example.com/grid-sim-game/game/systems/wiring"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// PlacementSystem is the only system that reads mouse clicks and turns them
// into toolbar tool selection, cell selection (when no tool is active), or
// grid placement. It mutates PlacementState and GridOccupancy; electrical
// join/leave is delegated to package wiring.
type PlacementSystem struct{}

// NewPlacementSystem builds the system. It holds no per-World state itself
// (everything it needs lives in resources), so w is unused; it is accepted
// for consistency with the engine's "build systems from a World" convention.
func NewPlacementSystem(_ *ecs.World) *PlacementSystem {
	return &PlacementSystem{}
}

// Update refreshes hover every frame, clears the tool on C, and handles at
// most one left-click (toolbar, select, or place/delete).
func (s *PlacementSystem) Update(w *ecs.World, dt float64) {
	in := ecs.GetResource[components.Input](w)
	placement := ecs.GetResource[grid.PlacementState](w)
	occupancy := ecs.GetResource[grid.GridOccupancy](w)
	cam := ecs.GetResource[components.Camera](w)
	if in == nil || placement == nil || occupancy == nil || cam == nil {
		return
	}

	mouse := in.State.Mouse
	bounds := ecs.GetResource[components.ScreenBounds](w)
	s.updateHover(placement, cam, bounds, mouse.X, mouse.Y)

	// C clears the active tool (and cancels a pending line); selection stays.
	if in.State.CPressed && !in.State.CPressedLastFrame {
		placement.Tool = grid.ToolNone
		placement.LinePending = false
	}

	if !mouse.Left.Pressed || mouse.Left.PressedLastFrame {
		return
	}

	// Right half is the ImGui network panel — ignore clicks there.
	if bounds != nil && mouse.X >= gameconfig.Global.PlayfieldWidth(bounds.W) {
		return
	}

	if mouse.Y < gameconfig.Global.ToolbarHeight {
		s.handleToolbarClick(placement, mouse.X, mouse.Y)
		return
	}

	if placement.Tool == grid.ToolNone {
		if placement.HoverValid {
			placement.HasSelection = true
			placement.SelectedCell = placement.HoverCell
		}
		return
	}

	s.handleGridClick(w, placement, occupancy, cam, mouse.X, mouse.Y)
}

// updateHover sets HoverCell / HoverValid from the cursor. Invalid over the
// toolbar, ImGui panel, or outside the grid.
func (s *PlacementSystem) updateHover(placement *grid.PlacementState, cam *components.Camera, bounds *components.ScreenBounds, x, y float64) {
	placement.HoverValid = false
	if bounds != nil && x >= gameconfig.Global.PlayfieldWidth(bounds.W) {
		return
	}
	if y < gameconfig.Global.ToolbarHeight {
		return
	}
	cell, ok := grid.ScreenToCell(cam, x, y)
	if !ok {
		return
	}
	placement.HoverCell = cell
	placement.HoverValid = true
}

func (s *PlacementSystem) handleToolbarClick(placement *grid.PlacementState, x, y float64) {
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
		if e, ok := s.placeSingle(w, occupancy, cell, grid.SpawnGenerator); ok {
			wiring.Attach(w, e, grid.ToolGenerator, cell, occupancy)
		}
	case grid.ToolHouse:
		if e, ok := s.placeSingle(w, occupancy, cell, grid.SpawnHouse); ok {
			wiring.Attach(w, e, grid.ToolHouse, cell, occupancy)
		}
	case grid.ToolDelete:
		deleteCell(w, occupancy, cell)
	case grid.ToolLine:
		s.handleLineClick(w, placement, occupancy, cell)
	}
}

func (s *PlacementSystem) placeSingle(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord, spawn func(*ecs.World, grid.GridCoord) ecs.Entity) (ecs.Entity, bool) {
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
			continue
		}
		e := grid.SpawnLineSegment(w, c)
		occupancy.Occupy(c, e)
		wiring.Attach(w, e, grid.ToolLine, c, occupancy)
	}
}

// deleteCell removes the entity at cell from the grid occupancy, the
// ElectricalNetwork (via wiring.Detach), and the ECS world.
// Returns true if something was deleted, false if the cell was empty.
func deleteCell(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord) bool {
	e, ok := occupancy.Cells[cell]
	if !ok {
		logger.Logger.Debugf("grid-sim: delete on empty cell %+v, ignoring", cell)
		return false
	}
	delete(occupancy.Cells, cell)
	wiring.Detach(w, e)
	w.Remove(e)
	return true
}

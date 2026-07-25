// Package placement implements PlacementSystem, which translates mouse clicks
// into toolbar tool selection and grid entity spawning.
package placement

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
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

	if mouse.Y < gameconfig.Global.ToolbarHeight {
		s.handleToolbarClick(placement, mouse.X, mouse.Y)
		return
	}
	s.handleGridClick(w, placement, occupancy, cam, mouse.X, mouse.Y)
}

// handleToolbarClick selects the clicked button's tool, or deselects it
// (ToolNone) if it was already selected. Any pending line is cancelled
// whenever the tool changes.
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

// handleGridClick converts a click below the toolbar into a grid cell (via
// the current camera offset) and dispatches to the placement logic for the
// currently selected tool.
func (s *PlacementSystem) handleGridClick(w *ecs.World, placement *grid.PlacementState, occupancy *grid.GridOccupancy, cam *components.Camera, x, y float64) {
	ts := gameconfig.Global.TileSize
	cell := grid.GridCoord{
		Col: int((x + cam.X) / ts),
		Row: int((y + cam.Y) / ts),
	}
	if cell.Col < 0 || cell.Row < 0 || cell.Col >= gameconfig.Global.GridCols || cell.Row >= gameconfig.Global.GridRows {
		return
	}

	switch placement.Tool {
	case grid.ToolGenerator:
		if e, ok := s.placeSingle(w, occupancy, cell, grid.SpawnGenerator); ok {
			attachToNetwork(w, e, grid.ToolGenerator, cell, occupancy)
		}
	case grid.ToolHouse:
		if e, ok := s.placeSingle(w, occupancy, cell, grid.SpawnHouse); ok {
			attachToNetwork(w, e, grid.ToolHouse, cell, occupancy)
		}
	case grid.ToolLine:
		s.handleLineClick(w, placement, occupancy, cell)
	case grid.ToolDelete:
		deleteCell(w, occupancy, cell)
	}
}

// placeSingle spawns a one-cell entity via spawn if cell is free, returning
// the new entity and true on success, or the zero entity and false if occupied.
func (s *PlacementSystem) placeSingle(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord, spawn func(*ecs.World, grid.GridCoord) ecs.Entity) (ecs.Entity, bool) {
	if occupancy.Occupied(cell) {
		logger.Logger.Debugf("grid-sim: cell %+v is occupied, ignoring placement", cell)
		return ecs.Entity{}, false
	}
	e := spawn(w, cell)
	occupancy.Occupy(cell, e)
	return e, true
}

// handleLineClick implements the two-click line flow: the first click on a
// free cell records it as the pending start; the second computes the
// Manhattan path to it and, if every cell on that path is free, spawns a
// line-tile entity per cell and wires each into the network. An invalid
// second click (blocked path) cancels the pending line.
func (s *PlacementSystem) handleLineClick(w *ecs.World, placement *grid.PlacementState, occupancy *grid.GridOccupancy, cell grid.GridCoord) {
	if !placement.LinePending {
		if occupancy.Occupied(cell) {
			return
		}
		placement.LineStart = cell
		placement.LinePending = true
		return
	}

	placement.LinePending = false

	path := grid.ManhattanPath(placement.LineStart, cell)
	for _, c := range path {
		if occupancy.Occupied(c) {
			logger.Logger.Debugf("grid-sim: line path blocked at %+v, cancelling", c)
			return
		}
	}
	for _, c := range path {
		e := grid.SpawnLineSegment(w, c)
		occupancy.Occupy(c, e)
		attachToNetwork(w, e, grid.ToolLine, c, occupancy)
	}
}

// deleteCell removes the entity at cell from the grid occupancy, the
// ElectricalNetwork (cascading to all incident branches), and the ECS world.
// Returns true if something was deleted, false if the cell was empty.
func deleteCell(w *ecs.World, occupancy *grid.GridOccupancy, cell grid.GridCoord) bool {
	e, ok := occupancy.Cells[cell]
	if !ok {
		logger.Logger.Debugf("grid-sim: delete on empty cell %+v, ignoring", cell)
		return false
	}
	delete(occupancy.Cells, cell)

	if net := ecs.GetResource[network.ElectricalNetwork](w); net != nil {
		if bus, ok := net.BusForEntity(e); ok {
			net.RemoveBus(bus.ID)
		}
	}

	w.Remove(e)
	return true
}

// attachToNetwork registers a freshly spawned entity as a bus in the
// ElectricalNetwork, stamps a NetworkLink component on the entity, then adds
// branches to any already-placed cardinal neighbours that are also in the
// network. It is a no-op if the ElectricalNetwork resource is absent.
//
// For line-segment entities the branch resistance is read from the entity's
// LineSegmentProps component; all other connections use resistance = 0.
// House entities have their HouseLoad demand written into the bus's BusSpec
// (consumer kW/kVAR → generator-convention watts/VAR).
//
// Graph mutations mark the network Dirty; LoadflowSystem re-solves later.
func attachToNetwork(w *ecs.World, e ecs.Entity, kind grid.Tool, cell grid.GridCoord, occupancy *grid.GridOccupancy) {
	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil {
		return
	}
	bus := net.AddBus(e, toolToBusType(kind))
	ecs.NewMap1[network.NetworkLink](w).Add(e, &network.NetworkLink{BusID: bus.ID})

	if kind == grid.ToolHouse {
		if hl := ecs.NewMap1[grid.HouseLoad](w).Get(e); hl != nil {
			net.SetBusSpec(bus.ID, network.PQSpec(-hl.PKw*1000, -hl.QKw*1000))
		}
	}

	var resistance float64
	if kind == grid.ToolLine {
		if lsp := ecs.NewMap1[grid.LineSegmentProps](w).Get(e); lsp != nil {
			resistance = lsp.ResistanceOhm
		}
	}

	dirs := []grid.GridCoord{{Col: 1}, {Col: -1}, {Row: 1}, {Row: -1}}
	for _, d := range dirs {
		nb := grid.GridCoord{Col: cell.Col + d.Col, Row: cell.Row + d.Row}
		ne, ok := occupancy.Cells[nb]
		if !ok {
			continue
		}
		if nbBus, ok := net.BusForEntity(ne); ok {
			net.AddBranch(bus.ID, nbBus.ID, resistance)
		}
	}
}

func toolToBusType(t grid.Tool) network.BusType {
	switch t {
	case grid.ToolGenerator:
		return network.BusGenerator
	case grid.ToolHouse:
		return network.BusLoad
	default:
		return network.BusJunction
	}
}

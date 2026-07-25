// Package wiring joins freshly placed grid entities into ElectricalNetwork
// (and tears them out on delete). Placement owns input/occupancy; this package
// owns bus/branch/history mutations so placement stays input-focused.
package wiring

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// Attach registers entity e as a bus, stamps NetworkLink + BusHistory, writes
// house PQ when applicable, and adds branches to network neighbours.
// Lines use AttachLine (one bus for the whole polyline). Graph mutations mark
// the network Dirty; LoadflowSystem re-solves later.
func Attach(w *ecs.World, e ecs.Entity, kind grid.Tool, cell grid.GridCoord, occupancy *grid.GridOccupancy) {
	if kind == grid.ToolLine {
		AttachLine(w, e, occupancy)
		return
	}

	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil {
		return
	}
	bus, err := net.AddBus(e, toolToBusType(kind))
	if err != nil {
		logger.Logger.Errorf("grid-sim: Attach AddBus: %v", err)
		return
	}
	ecs.NewMap1[network.NetworkLink](w).Add(e, &network.NetworkLink{BusID: bus.ID})
	h := network.NewBusHistory()
	ecs.NewMap1[network.BusHistory](w).Add(e, &h)

	if kind == grid.ToolHouse {
		if hl := ecs.NewMap1[grid.HouseLoad](w).Get(e); hl != nil {
			net.SetBusSpec(bus.ID, network.PQSpec(-hl.PKw*1000, -hl.QKw*1000))
		}
	}

	// Contact links (R=X=0) to non-line neighbours; line neighbours are
	// rewired below so stroke impedance stays on the line spokes.
	dirs := []grid.GridCoord{{Col: 1}, {Col: -1}, {Row: 1}, {Row: -1}}
	var lineNeighbours []ecs.Entity
	for _, d := range dirs {
		nb := grid.GridCoord{Col: cell.Col + d.Col, Row: cell.Row + d.Row}
		ne, ok := occupancy.Cells[nb]
		if !ok {
			continue
		}
		if ecs.NewMap1[grid.LinePath](w).Get(ne) != nil || ecs.NewMap1[grid.LineSegmentProps](w).Get(ne) != nil {
			lineNeighbours = append(lineNeighbours, ne)
			continue
		}
		if nbBus, ok := net.BusForEntity(ne); ok {
			net.AddBranch(bus.ID, nbBus.ID, 0, 0)
		}
	}
	for _, le := range uniqueEntities(lineNeighbours) {
		rewireLineSpokes(w, net, le, occupancy)
	}
}

// AttachLine registers a polyline line entity as one Junction bus and wires
// spokes to neighbouring network buses (see rewireLineSpokes).
func AttachLine(w *ecs.World, e ecs.Entity, occupancy *grid.GridOccupancy) {
	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil {
		return
	}
	bus, err := net.AddBus(e, network.BusJunction)
	if err != nil {
		logger.Logger.Errorf("grid-sim: AttachLine AddBus: %v", err)
		return
	}
	ecs.NewMap1[network.NetworkLink](w).Add(e, &network.NetworkLink{BusID: bus.ID})
	h := network.NewBusHistory()
	ecs.NewMap1[network.BusHistory](w).Add(e, &h)
	rewireLineSpokes(w, net, e, occupancy)
}

// Detach removes the entity's bus (and incident branches) from the network if
// linked. Caller owns GridOccupancy and ECS entity removal.
func Detach(w *ecs.World, e ecs.Entity) {
	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil {
		return
	}
	if bus, ok := net.BusForEntity(e); ok {
		net.RemoveBus(bus.ID)
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

func rewireLineSpokes(w *ecs.World, net *network.ElectricalNetwork, lineEntity ecs.Entity, occupancy *grid.GridOccupancy) {
	bus, ok := net.BusForEntity(lineEntity)
	if !ok {
		return
	}
	// Drop existing spokes; rebuild from current occupancy.
	for _, brID := range append([]network.BranchID(nil), incidentCopy(net, bus.ID)...) {
		net.RemoveBranch(brID)
	}

	var r, x float64
	if lsp := ecs.NewMap1[grid.LineSegmentProps](w).Get(lineEntity); lsp != nil {
		r, x = lsp.ResistanceOhm, lsp.ReactanceOhm
	}
	neighbors := uniqueNeighborBuses(net, lineEntity, lineCells(w, lineEntity), occupancy)
	n := len(neighbors)
	if n == 0 {
		return
	}
	sr, sx := r/float64(n), x/float64(n)
	for _, nbID := range neighbors {
		net.AddBranch(bus.ID, nbID, sr, sx)
	}
}

func incidentCopy(net *network.ElectricalNetwork, id network.BusID) []network.BranchID {
	// Neighbors walks incidence; we need branch IDs. Use Branches() filter.
	var ids []network.BranchID
	for brID, br := range net.Branches() {
		if br.From == id || br.To == id {
			ids = append(ids, brID)
		}
	}
	return ids
}

func lineCells(w *ecs.World, e ecs.Entity) []grid.GridCoord {
	if lp := ecs.NewMap1[grid.LinePath](w).Get(e); lp != nil && len(lp.Cells) > 0 {
		return lp.Cells
	}
	if go_ := ecs.NewMap1[grid.GridObject](w).Get(e); go_ != nil {
		return []grid.GridCoord{go_.Cell}
	}
	return nil
}

func uniqueNeighborBuses(
	net *network.ElectricalNetwork,
	self ecs.Entity,
	cells []grid.GridCoord,
	occupancy *grid.GridOccupancy,
) []network.BusID {
	seen := make(map[network.BusID]bool)
	var out []network.BusID
	dirs := []grid.GridCoord{{Col: 1}, {Col: -1}, {Row: 1}, {Row: -1}}
	for _, cell := range cells {
		for _, d := range dirs {
			nb := grid.GridCoord{Col: cell.Col + d.Col, Row: cell.Row + d.Row}
			ne, ok := occupancy.Cells[nb]
			if !ok || ne == self {
				continue
			}
			nbBus, ok := net.BusForEntity(ne)
			if !ok || seen[nbBus.ID] {
				continue
			}
			seen[nbBus.ID] = true
			out = append(out, nbBus.ID)
		}
	}
	return out
}

func uniqueEntities(in []ecs.Entity) []ecs.Entity {
	seen := make(map[ecs.Entity]bool)
	var out []ecs.Entity
	for _, e := range in {
		if seen[e] {
			continue
		}
		seen[e] = true
		out = append(out, e)
	}
	return out
}

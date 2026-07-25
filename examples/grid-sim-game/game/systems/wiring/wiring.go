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
// house PQ when applicable, and adds branches to cardinal network neighbours.
// Branch resistance comes from LineSegmentProps for line tiles; otherwise 0.
// Graph mutations mark the network Dirty; LoadflowSystem re-solves later.
func Attach(w *ecs.World, e ecs.Entity, kind grid.Tool, cell grid.GridCoord, occupancy *grid.GridOccupancy) {
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

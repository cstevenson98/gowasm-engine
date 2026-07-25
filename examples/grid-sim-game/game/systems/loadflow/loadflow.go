// Package loadflow implements LoadflowSystem: when the ElectricalNetwork
// resource is marked Dirty (topology or bus-spec change), it runs one AC
// power-flow solve and logs the resulting bus voltages. On frames where the
// circuit is unchanged it is a no-op.
package loadflow

import (
	"example.com/grid-sim-game/game/components/network"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
)

// LoadflowSystem owns network analysis. Placement (and any other mutator)
// only touches the ElectricalNetwork graph/specs and leaves Dirty set;
// this system is the sole caller of LoadflowSolver.Solve.
type LoadflowSystem struct {
	solver *network.LoadflowSolver
}

// NewLoadflowSystem builds the system with a default LoadflowSolver.
func NewLoadflowSystem(_ *ecs.World) *LoadflowSystem {
	return &LoadflowSystem{solver: network.NewLoadflowSolver()}
}

// Update re-solves the network if Dirty, then clears the flag. A failed
// solve still clears Dirty so we do not spam Newton-Raphson every frame
// for an unchanged (e.g. no-slack) circuit; the next mutation will mark
// Dirty again.
func (s *LoadflowSystem) Update(w *ecs.World, _ float64) {
	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil || !net.Dirty {
		return
	}

	net.Log()
	if err := s.solver.Solve(net); err != nil {
		logger.Logger.Errorf("grid-sim: loadflow failed: %v", err)
	}
	net.LogVoltages()
	net.ClearDirty()
}

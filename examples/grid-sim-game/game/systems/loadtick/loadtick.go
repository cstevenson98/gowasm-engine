// Package loadtick implements LoadTickSystem, which periodically re-samples
// house demand and pushes the new P/Q into ElectricalNetwork bus specs so
// LoadflowSystem re-solves on the next dirty pass.
package loadtick

import (
	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// DefaultInterval is how often house loads are re-sampled.
const DefaultInterval = 3.0 // seconds

// LoadTickSystem accumulates dt and, every Interval seconds, assigns a new
// random P/Q to every HouseLoad entity that is linked into the network, then
// updates the corresponding BusSpec (which marks the network Dirty).
type LoadTickSystem struct {
	Interval float64
	acc      float64
	houses   *ecs.Filter2[grid.HouseLoad, network.NetworkLink]
}

// NewLoadTickSystem builds the system with DefaultInterval.
func NewLoadTickSystem(w *ecs.World) *LoadTickSystem {
	return &LoadTickSystem{
		Interval: DefaultInterval,
		houses:   ecs.NewFilter2[grid.HouseLoad, network.NetworkLink](w),
	}
}

// Update advances the timer; when Interval elapses, all houses are re-sampled.
func (s *LoadTickSystem) Update(w *ecs.World, dt float64) {
	if s.Interval <= 0 {
		s.Interval = DefaultInterval
	}
	s.acc += dt
	if s.acc < s.Interval {
		return
	}
	s.acc = 0

	net := ecs.GetResource[network.ElectricalNetwork](w)
	if net == nil {
		return
	}

	s.houses.Each(func(_ ecs.Entity, hl *grid.HouseLoad, link *network.NetworkLink) {
		hl.PKw = grid.RandLoadKW()
		hl.QKw = grid.RandLoadKW()
		// Consumer kW → generator-convention watts (negative = load).
		net.SetBusSpec(link.BusID, network.PQSpec(-hl.PKw*1000, -hl.QKw*1000))
	})
}

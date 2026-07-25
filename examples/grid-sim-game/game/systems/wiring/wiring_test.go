package wiring_test

import (
	"testing"

	"example.com/grid-sim-game/game/components/grid"
	"example.com/grid-sim-game/game/components/network"
	"example.com/grid-sim-game/game/systems/loadflow"
	"example.com/grid-sim-game/game/systems/loadtick"
	"example.com/grid-sim-game/game/systems/wiring"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

func TestAttachDetachAndLoadflowPipeline(t *testing.T) {
	w := ecs.NewWorld()
	occ := grid.NewGridOccupancy()
	net := network.NewElectricalNetwork()
	ecs.SetResource(w, occ)
	ecs.SetResource(w, net)

	c0 := grid.GridCoord{Col: 0, Row: 0}
	c1 := grid.GridCoord{Col: 1, Row: 0}
	c2 := grid.GridCoord{Col: 2, Row: 0}

	gen := grid.SpawnGenerator(w, c0)
	line := grid.SpawnLineSegment(w, c1)
	house := grid.SpawnHouse(w, c2)
	occ.Occupy(c0, gen)
	occ.Occupy(c1, line)
	occ.Occupy(c2, house)

	wiring.Attach(w, gen, grid.ToolGenerator, c0, occ)
	wiring.Attach(w, line, grid.ToolLine, c1, occ)
	wiring.Attach(w, house, grid.ToolHouse, c2, occ)

	if len(net.Buses()) != 3 {
		t.Fatalf("buses = %d, want 3", len(net.Buses()))
	}
	if len(net.Branches()) != 2 {
		t.Fatalf("branches = %d, want 2", len(net.Branches()))
	}
	if !net.Dirty {
		t.Fatal("expected Dirty after Attach")
	}

	lf := loadflow.NewLoadflowSystem(w)
	lf.Update(w, 0)
	if net.Dirty {
		t.Fatal("Dirty should clear after loadflow")
	}
	if !net.State.Converged {
		t.Fatalf("expected converged solve, iters=%d", net.State.Iterations)
	}

	// Force a load tick by setting a tiny interval and large dt.
	tick := loadtick.NewLoadTickSystem(w)
	tick.Interval = 1
	tick.Update(w, 2)
	if !net.Dirty {
		t.Fatal("expected Dirty after load tick")
	}

	lf.Update(w, 0)
	if net.Dirty {
		t.Fatal("Dirty should clear after second loadflow")
	}

	wiring.Detach(w, house)
	w.Remove(house)
	delete(occ.Cells, c2)
	if _, ok := net.BusForEntity(house); ok {
		t.Fatal("house bus should be gone after Detach")
	}
	if len(net.Buses()) != 2 {
		t.Fatalf("buses = %d after detach, want 2", len(net.Buses()))
	}
}

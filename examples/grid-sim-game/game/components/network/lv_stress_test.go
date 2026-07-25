package network_test

import (
	"testing"

	"example.com/grid-sim-game/game/components/network"
)

// TestLVZeroContactR mirrors in-game branch resistances: line→gen gets the
// segment R (0.05Ω), house→line gets 0 (direct contact, clamped to
// minResistance). This used to stall Newton at ~1e-6 residual because the
// 1e-6Ω floor made Y-bus entries ~1e6 S in SI units.
func TestLVZeroContactR(t *testing.T) {
	net, w := emptyNet()
	e0, e1, e2 := newEntity(w), newEntity(w), newEntity(w)
	b0 := net.AddBus(e0, network.BusGenerator)
	b1 := net.AddBus(e1, network.BusJunction)
	b2 := net.AddBus(e2, network.BusLoad)
	net.SetBusSpec(b2.ID, network.PQSpec(-17390, -15000))
	net.AddBranch(b0.ID, b1.ID, 0.05)
	net.AddBranch(b1.ID, b2.ID, 0) // in-game house→line contact

	if err := network.NewLoadflowSolver().Solve(net); err != nil {
		t.Fatalf("solve error: %v", err)
	}
	if !net.State.Converged {
		t.Fatalf("did not converge, iterations=%d", net.State.Iterations)
	}

	bs1, _ := net.BusStateFor(b1.ID)
	bs2, _ := net.BusStateFor(b2.ID)
	// Contact branch is minResistance (1e-3 Ω): expect sub-volt drop at ~75 A.
	if d := abs(bs1.Result.VoltMag - bs2.Result.VoltMag); d > 0.2 {
		t.Errorf("|V1-V2|=%.4f, want nearly equal for contact link", d)
	}
	t.Logf("converged in %d iters: V1=%.2f V2=%.2f", net.State.Iterations, bs1.Result.VoltMag, bs2.Result.VoltMag)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

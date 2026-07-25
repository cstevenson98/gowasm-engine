package network_test

import (
	"math"
	"testing"

	"example.com/grid-sim-game/game/components/network"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// helper: spawn a dummy ECS entity so we can attach a bus to it.
func newEntity(w *ecs.World) ecs.Entity {
	return ecs.NewMap1[struct{}](w).NewEntity(&struct{}{})
}

// helper: build a fresh network + world, return them.
func emptyNet() (*network.ElectricalNetwork, *ecs.World) {
	return network.NewElectricalNetwork(), ecs.NewWorld()
}

// TestFlatStart: a network with only a slack bus should converge immediately
// with no mismatch.
func TestFlatStart(t *testing.T) {
	net, w := emptyNet()
	e0 := newEntity(w)
	bus0 := net.AddBus(e0, network.BusGenerator)
	net.SetBusSpec(bus0.ID, network.SlackSpec(1.0, 0.0))

	s := network.NewLoadflowSolver()
	if err := s.Solve(net); err != nil {
		t.Fatalf("solve error: %v", err)
	}
	if !net.State.Converged {
		t.Fatal("expected convergence for trivial slack-only network")
	}

	bs, _ := net.BusStateFor(bus0.ID)
	if math.Abs(bs.Result.VoltMag-1.0) > 1e-9 {
		t.Errorf("|V| = %v, want 1.0", bs.Result.VoltMag)
	}
}

// TestTwoBusPureResistive verifies the NR solver against the analytical
// solution of a 2-bus network with one slack bus and one PQ load bus,
// connected by a purely resistive branch R = 1 pu.
//
// Network:
//
//	Bus 0 (Slack, V=1∠0°) ── R=1 ── Bus 1 (PQ, P_spec=−0.05, Q_spec=0)
//
// Generator convention: P_spec < 0 = absorbing from the network.
//
// Analytical solution (from the DC power equation for θ=0, B=0):
//
//	P_1 = V_1·(V_1 − 1) = −0.05
//	V_1² − V_1 + 0.05 = 0
//	V_1 = (1 + √0.8) / 2  ≈  0.9472    (high-voltage root)
//	δ_1 = 0   (no reactive power → no angle, from Q_1 = 0)
func TestTwoBusPureResistive(t *testing.T) {
	const R = 1.0
	const Pspec = -0.05
	wantV := (1.0 + math.Sqrt(0.8)) / 2.0 // ≈ 0.9472

	net, w := emptyNet()
	e0 := newEntity(w)
	e1 := newEntity(w)

	bus0 := net.AddBus(e0, network.BusGenerator)
	net.SetBusSpec(bus0.ID, network.SlackSpec(1.0, 0.0))

	bus1 := net.AddBus(e1, network.BusLoad)
	net.SetBusSpec(bus1.ID, network.PQSpec(Pspec, 0.0))

	net.AddBranch(bus0.ID, bus1.ID, R)

	s := network.NewLoadflowSolver()
	if err := s.Solve(net); err != nil {
		t.Fatalf("solve error: %v", err)
	}
	if !net.State.Converged {
		t.Fatalf("did not converge, iterations=%d", net.State.Iterations)
	}

	bs1, _ := net.BusStateFor(bus1.ID)
	if math.Abs(bs1.Result.VoltMag-wantV) > 1e-6 {
		t.Errorf("|V_1| = %.8f, want %.8f (err=%.2e)", bs1.Result.VoltMag, wantV, math.Abs(bs1.Result.VoltMag-wantV))
	}
	if math.Abs(bs1.Result.VoltAng) > 1e-6 {
		t.Errorf("δ_1 = %.2e, want 0", bs1.Result.VoltAng)
	}
	t.Logf("converged in %d iterations: |V_1|=%.8f δ_1=%.4f°",
		net.State.Iterations, bs1.Result.VoltMag, bs1.Result.VoltAng*180/math.Pi)

	// Verify P_calc at bus 1 matches P_spec.
	if math.Abs(bs1.Result.PInject-Pspec) > 1e-6 {
		t.Errorf("P_calc_1 = %.8f, want %.8f", bs1.Result.PInject, Pspec)
	}
}

// TestThreeBusResistive checks a simple three-bus radial network:
//
//	Bus 0 (Slack, V=1∠0°) ── R=0.1 ── Bus 1 (PQ, P=−0.02, Q=0) ── R=0.1 ── Bus 2 (PQ, P=−0.02, Q=0)
//
// We check that:
//  1. The solver converges.
//  2. V_2 < V_1 < V_0 (voltage decreases along the feeder).
//  3. P_calc matches P_spec at buses 1 and 2 to within tolerance.
func TestThreeBusResistive(t *testing.T) {
	const R = 0.1
	const Pload = -0.02

	net, w := emptyNet()
	e0, e1, e2 := newEntity(w), newEntity(w), newEntity(w)

	b0 := net.AddBus(e0, network.BusGenerator)
	net.SetBusSpec(b0.ID, network.SlackSpec(1.0, 0.0))

	b1 := net.AddBus(e1, network.BusLoad)
	net.SetBusSpec(b1.ID, network.PQSpec(Pload, 0.0))

	b2 := net.AddBus(e2, network.BusLoad)
	net.SetBusSpec(b2.ID, network.PQSpec(Pload, 0.0))

	net.AddBranch(b0.ID, b1.ID, R)
	net.AddBranch(b1.ID, b2.ID, R)

	s := network.NewLoadflowSolver()
	if err := s.Solve(net); err != nil {
		t.Fatalf("solve error: %v", err)
	}
	if !net.State.Converged {
		t.Fatalf("did not converge, iterations=%d", net.State.Iterations)
	}

	bs0, _ := net.BusStateFor(b0.ID)
	bs1, _ := net.BusStateFor(b1.ID)
	bs2, _ := net.BusStateFor(b2.ID)

	t.Logf("converged in %d iters: V0=%.6f V1=%.6f V2=%.6f",
		net.State.Iterations, bs0.Result.VoltMag, bs1.Result.VoltMag, bs2.Result.VoltMag)

	if bs1.Result.VoltMag >= bs0.Result.VoltMag {
		t.Errorf("expected V_1 < V_0, got V_1=%.6f V_0=%.6f", bs1.Result.VoltMag, bs0.Result.VoltMag)
	}
	if bs2.Result.VoltMag >= bs1.Result.VoltMag {
		t.Errorf("expected V_2 < V_1, got V_2=%.6f V_1=%.6f", bs2.Result.VoltMag, bs1.Result.VoltMag)
	}

	const tol = 1e-5
	for _, tc := range []struct {
		name  string
		bs    *network.BusState
		pSpec float64
	}{
		{"bus1", bs1, Pload},
		{"bus2", bs2, Pload},
	} {
		if math.Abs(tc.bs.Result.PInject-tc.pSpec) > tol {
			t.Errorf("%s: P_calc=%.8f, want %.8f", tc.name, tc.bs.Result.PInject, tc.pSpec)
		}
	}
}

// TestNoSlackBusReturnsError ensures Solve returns an error when no slack bus
// is present (the system would be singular).
func TestNoSlackBusReturnsError(t *testing.T) {
	net, w := emptyNet()
	e := newEntity(w)
	b := net.AddBus(e, network.BusLoad)
	net.SetBusSpec(b.ID, network.PQSpec(-0.1, 0))

	s := network.NewLoadflowSolver()
	if err := s.Solve(net); err == nil {
		t.Fatal("expected error for network with no slack bus")
	}
}

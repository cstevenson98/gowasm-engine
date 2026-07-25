package network

import "fmt"

// Solver is the interface for network analysis algorithms. Implementations
// read BusSpec boundary conditions from net.State and write BusResult /
// BranchResult solved quantities back into the same StaticState.
type Solver interface {
	Solve(net *ElectricalNetwork) error
}

// LoadflowSolver performs a static AC load flow. The current implementation
// is a flat-start stub: it writes V=1.0∠0° to every bus and zero flow to
// every branch, then marks the state as converged. Replace the body of Solve
// with Newton-Raphson or a DC approximation when ready.
type LoadflowSolver struct{}

// NewLoadflowSolver creates a LoadflowSolver.
func NewLoadflowSolver() *LoadflowSolver { return &LoadflowSolver{} }

// Solve writes flat-start values to net.State.
func (s *LoadflowSolver) Solve(net *ElectricalNetwork) error {
	if net.State == nil {
		return fmt.Errorf("loadflow: network has no StaticState")
	}
	st := net.State

	for _, bs := range st.Buses {
		switch bs.Spec.Formulation {
		case Slack:
			// V and θ are fixed; stub copies them straight through.
			// A real solver would compute P and Q at this bus.
			bs.Result = BusResult{
				VoltMag: bs.Spec.VoltMag,
				VoltAng: bs.Spec.VoltAng,
			}
		case PV:
			// V and P are fixed; stub keeps the voltage, leaves θ = 0.
			// A real solver would compute θ and Q.
			bs.Result = BusResult{
				VoltMag: bs.Spec.VoltMag,
				PInject: bs.Spec.PInject,
			}
		case PQ:
			// P and Q are fixed; stub uses flat-start V = 1.0∠0°.
			// A real solver would compute V and θ.
			bs.Result = BusResult{
				VoltMag: 1.0,
				PInject: bs.Spec.PInject,
				QInject: bs.Spec.QInject,
			}
		}
	}
	for _, br := range st.Branches {
		br.Result = BranchResult{}
	}

	st.Converged = true
	st.Iterations = 0
	return nil
}

// TimeEvolution holds a series of StaticState snapshots and advances the
// network through time by repeatedly solving and recording results.
// Not yet implemented — see plans/loadflow.md.
type TimeEvolution struct {
	History []*StaticState
	dt      float64 // time step in seconds
}

// NewTimeEvolution creates an empty time evolution with the given step size.
func NewTimeEvolution(dt float64) *TimeEvolution {
	return &TimeEvolution{dt: dt}
}

// Step updates bus specs from the network's current entity state, solves,
// and appends a copy of the result to History.
func (t *TimeEvolution) Step(_ *ElectricalNetwork, _ Solver) error {
	panic("TimeEvolution.Step: not implemented")
}

//go:build !cgo

// Fallback for builds without CGo support — notably GOOS=js GOARCH=wasm
// (the grid-sim-game browser build), where CGO_ENABLED is forced to 0 and
// superlu_cgo.go is excluded. Callers that need a working solver on such
// builds should use SparseLUSolver instead.
package nr

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// SuperLUSolver is unavailable without CGo. It returns a LinearSolver that
// always errors, so a caller which unconditionally wires up SuperLUSolver()
// fails loudly (at solve time) rather than silently producing wrong results.
func SuperLUSolver() LinearSolver {
	return func(_ mat.Matrix, _ *mat.VecDense) (*mat.VecDense, error) {
		return nil, fmt.Errorf("superlu: not available in this build (CGo disabled); use SparseLUSolver instead")
	}
}

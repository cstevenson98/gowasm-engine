// Package nr implements a generic Newton-Raphson root-finder.
//
// It is completely independent of power-flow or any other physical problem.
// Callers supply two closures:
//
//	f(x)   — the residual vector; the solver drives ‖f(x)‖₂ → 0
//	jac(x) — the Jacobian matrix ∂f/∂x at the current iterate
//
// The linear solve at each step is handled by a pluggable LinearSolver.
// The default, SparseLUSolver, accepts any mat.Matrix (including a sparse
// CSR from james-bowman/sparse) and solves via LU factorisation. For systems
// up to ~800×800 (a 400-bus full AC power flow) this is adequate; swap in a
// fill-reducing sparse direct solver (e.g. AMD + UMFPACK via CGo) by passing
// a different LinearSolver to NewtonRaphson.LinearSolve.
package nr

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

// ResidualFunc computes the residual vector f(x).
type ResidualFunc func(x *mat.VecDense) *mat.VecDense

// JacobianFunc computes the Jacobian matrix J(x) = ∂f/∂x at x.
// Returning mat.Matrix allows callers to return a sparse.CSR, mat.Dense,
// or any other gonum-compatible matrix without changing this package.
type JacobianFunc func(x *mat.VecDense) mat.Matrix

// LinearSolver solves A·x = b and returns x. It is the seam between
// NewtonRaphson and the underlying linear algebra backend.
type LinearSolver func(A mat.Matrix, b *mat.VecDense) (*mat.VecDense, error)

// Result holds the outcome of a Newton-Raphson solve.
type Result struct {
	X          *mat.VecDense
	Iterations int
	Converged  bool
	Residual   float64 // ‖f(X)‖₂ at the final iterate
}

// NewtonRaphson drives  x_{k+1} = x_k - J(x_k)⁻¹ f(x_k)  to convergence.
type NewtonRaphson struct {
	MaxIter     int
	Tol         float64      // convergence threshold on ‖f(x)‖₂
	LinearSolve LinearSolver // defaults to SparseLUSolver() if nil
}

// Solve finds x* such that f(x*) ≈ 0, starting from x0.
// Returns a non-nil error if the linear solve fails or MaxIter is exhausted
// without convergence; the Result is always populated so callers can inspect
// the best iterate and residual even on failure.
func (nr *NewtonRaphson) Solve(
	x0 *mat.VecDense,
	f ResidualFunc,
	jac JacobianFunc,
) (*Result, error) {
	solve := nr.LinearSolve
	if solve == nil {
		solve = SparseLUSolver()
	}

	n := x0.Len()
	x := mat.NewVecDense(n, nil)
	x.CopyVec(x0)

	var residual float64
	for iter := 0; iter <= nr.MaxIter; iter++ {
		fx := f(x)
		residual = l2Norm(fx)

		if residual < nr.Tol {
			return &Result{X: x, Iterations: iter, Converged: true, Residual: residual}, nil
		}
		if iter == nr.MaxIter {
			break
		}

		J := jac(x)

		// Solve J·dx = -f(x)
		negFx := mat.NewVecDense(n, nil)
		negFx.ScaleVec(-1, fx)

		dx, err := solve(J, negFx)
		if err != nil {
			return &Result{X: x, Iterations: iter, Converged: false, Residual: residual},
				fmt.Errorf("nr: linear solve failed at iteration %d: %w", iter, err)
		}

		x.AddVec(x, dx)
	}

	return &Result{X: x, Iterations: nr.MaxIter, Converged: false, Residual: residual},
		fmt.Errorf("nr: did not converge after %d iterations (residual %.6e, tol %.6e)",
			nr.MaxIter, residual, nr.Tol)
}

// SparseLUSolver returns a LinearSolver that factorises A using LU and back-
// substitutes to find x in A·x = b.
//
// It accepts any mat.Matrix — including a sparse.CSR from
// github.com/james-bowman/sparse — because it only calls the mat.Matrix
// interface (At, Dims). Non-zero elements are copied into a gonum dense
// matrix before factorisation, so zero fill-in in the sparse input does not
// pay the copy cost.
//
// TODO: when the system grows beyond ~800×800, replace the inner copy with a
// fill-reducing sparse direct solve (e.g. convert to CSC → AMD reordering →
// UMFPACK via CGo, or a pure-Go nested-dissection implementation).
func SparseLUSolver() LinearSolver {
	return func(A mat.Matrix, b *mat.VecDense) (*mat.VecDense, error) {
		n, _ := A.Dims()

		// Copy sparse (or dense) input into a gonum dense matrix.
		// Only non-zero entries are written; zero entries in A are skipped.
		dense := mat.NewDense(n, n, nil)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if v := A.At(i, j); v != 0 {
					dense.Set(i, j, v)
				}
			}
		}

		var lu mat.LU
		lu.Factorize(dense)

		x := mat.NewVecDense(n, nil)
		if err := lu.SolveVecTo(x, false, b); err != nil {
			return nil, fmt.Errorf("sparse LU solve: %w", err)
		}
		return x, nil
	}
}

// l2Norm returns the Euclidean norm of v: sqrt(v·v).
func l2Norm(v *mat.VecDense) float64 {
	return math.Sqrt(mat.Dot(v, v))
}

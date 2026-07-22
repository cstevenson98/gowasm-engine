package ecs

import arkecs "github.com/mlange-42/ark/ecs"

// Comp identifies a component type for filter refinement (With / Without).
// Create one with C.
type Comp struct{ inner arkecs.Comp }

// C creates a Comp for component type T.
//
//	filter.With(ecs.C[Velocity]())
func C[T any]() Comp { return Comp{inner: arkecs.C[T]()} }

func toArkComps(comps []Comp) []arkecs.Comp {
	out := make([]arkecs.Comp, len(comps))
	for i, c := range comps {
		out[i] = c.inner
	}
	return out
}

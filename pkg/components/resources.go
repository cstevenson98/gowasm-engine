package components

import "github.com/cstevenson98/milo/pkg/types"

// ScreenBounds is a per-World singleton resource holding the virtual screen
// size. Systems (e.g. screen wrapping) read it via ecs.GetResource. Seeded by
// BaseState.Enter from the engine-provided dimensions.
type ScreenBounds struct {
	W, H float64
}

// Input is a per-World singleton resource holding the latest input snapshot.
// The engine refreshes State each frame before running the state's schedule, so
// input systems can read it via ecs.GetResource without bespoke wiring.
type Input struct {
	State types.InputState
}

package components

// ScreenBounds is a per-World singleton resource holding the virtual screen
// size. Systems (e.g. screen wrapping) read it via ecs.GetResource. Seeded by
// BaseState.Enter from the engine-provided dimensions.
type ScreenBounds struct {
	W, H float64
}

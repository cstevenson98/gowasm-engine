package types

// Vector2 is a 2D vector used throughout the engine to represent positions,
// sizes, velocities, and offsets. Coordinates are in virtual screen pixels
// unless a specific API documents otherwise.
type Vector2 struct {
	X float64
	Y float64
}

// UVRect is a normalised sub-rectangle of a texture, used to select which
// region of a texture (or sprite-sheet frame) to sample when drawing. All
// values are in the range 0.0 to 1.0, where (0,0) is the top-left of the
// texture and (1,1) is the bottom-right.
type UVRect struct {
	U float64 // Left (0.0 to 1.0)
	V float64 // Top (0.0 to 1.0)
	W float64 // Width (0.0 to 1.0)
	H float64 // Height (0.0 to 1.0)
}

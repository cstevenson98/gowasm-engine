package types

// Vector2 represents a 2D vector
type Vector2 struct {
	X float64
	Y float64
}

// UVRect represents UV coordinates for texture sampling
type UVRect struct {
	U float64 // Left (0.0 to 1.0)
	V float64 // Top (0.0 to 1.0)
	W float64 // Width (0.0 to 1.0)
	H float64 // Height (0.0 to 1.0)
}

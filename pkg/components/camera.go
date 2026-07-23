package components

// Camera is a per-World singleton resource describing the current view into
// the world: the world-space position of the viewport's top-left corner, and
// a zoom factor (1 = no zoom, >1 = zoomed in, <=0 is treated as 1). Systems
// that move the camera (e.g. CameraFollow) mutate it in place each frame; the
// renderer reads it to offset the Background and Entities layers (the UI layer
// is always drawn in screen space, ignoring the camera). Seeded by
// BaseState.Enter with an identity camera (0,0, zoom 1), so states that never
// touch it render exactly as if there were no camera at all.
type Camera struct {
	X, Y float64
	Zoom float64
}

// CameraTarget marks the entity that the CameraFollow system should center the
// camera on. At most one entity should carry it; if none do (or the World has
// no CameraFollow system registered), the camera simply stays wherever it was
// last left.
type CameraTarget struct{}

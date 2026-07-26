// Package components defines the pure-data components and shared resource types
// used by the ECS engine. Components carry no behaviour; systems (in the engine
// and games) operate on them. The only methods here are pure projections of a
// component's own data (e.g. Sprite.UV), never state mutation or game logic.
package components

import "github.com/cstevenson98/milo/pkg/types"

// Position is an entity's top-left position in virtual screen pixels.
// Replaces Mover position.
type Position struct{ X, Y float64 }

// Velocity is movement in pixels per second. Entities without a Velocity are
// stationary (the movement system simply never matches them). Replaces Mover
// velocity.
type Velocity struct{ DX, DY float64 }

// Wrap marks an entity for screen-edge wrapping against the ScreenBounds
// resource. SpriteW/SpriteH are the extents used to wrap just off-screen.
// Entities without a Wrap component do not wrap.
type Wrap struct{ SpriteW, SpriteH float64 }

// Sprite is the render-facing appearance of an entity: which texture (or sprite
// sheet) to draw, its display size, the sheet grid, the current frame, and
// visibility. It is plain data; the animation and render systems act on it.
type Sprite struct {
	TexturePath string
	Size        types.Vector2
	Columns     int // sheet columns (n); 1 for a single image
	Rows        int // sheet rows (m); 1 for a single image
	Frame       int // current frame index, advanced by the animation system
	Visible     bool
}

// TotalFrames returns the number of frames in the sheet (>= 1).
func (s *Sprite) TotalFrames() int {
	n := s.Columns * s.Rows
	if n < 1 {
		return 1
	}
	return n
}

// UV returns the UV rectangle for the current frame. Pure projection of the
// sprite's own fields; no state is modified.
func (s *Sprite) UV() types.UVRect {
	cols, rows := s.Columns, s.Rows
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	fw := 1.0 / float64(cols)
	fh := 1.0 / float64(rows)
	fx := s.Frame % cols
	fy := (s.Frame / cols) % rows
	return types.UVRect{
		U: float64(fx) * fw,
		V: float64(fy) * fh,
		W: fw,
		H: fh,
	}
}

// Animation drives frame cycling for a Sprite on the same entity. Entities with
// a Sprite but no Animation are static (e.g. backgrounds). Replaces the timing
// state previously inside SpriteSheet.
type Animation struct {
	FrameTime float64 // seconds per frame
	Elapsed   float64 // accumulated time toward the next frame
}

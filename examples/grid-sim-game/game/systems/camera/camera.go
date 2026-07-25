// Package camera implements CameraScrollSystem, a free-scrolling camera
// driven directly by player input (arrow keys / WASD / gamepad), clamped to
// the grid's world bounds.
package camera

import (
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// CameraScrollSystem moves the World's Camera resource continuously while an
// arrow key (or WASD, or gamepad d-pad/stick - all merged into Input.MoveX by
// the engine) is held, clamped so the viewport never scrolls past the grid's
// world bounds. Unlike systems.CameraFollow (which centers on an entity),
// this is a free-scrolling camera driven directly by input, so it lives here
// rather than in the engine's generic systems package.
type CameraScrollSystem struct {
	speed float64
}

// NewCameraScrollSystem builds the system with the given scroll speed (pixels
// per second).
func NewCameraScrollSystem(speed float64) *CameraScrollSystem {
	return &CameraScrollSystem{speed: speed}
}

// Update advances the camera and clamps it to the grid's world bounds.
func (s *CameraScrollSystem) Update(w *ecs.World, dt float64) {
	cam := ecs.GetResource[components.Camera](w)
	bounds := ecs.GetResource[components.ScreenBounds](w)
	if cam == nil || bounds == nil {
		return
	}

	if in := ecs.GetResource[components.Input](w); in != nil {
		st := in.State
		if st.MoveLeft {
			cam.X -= s.speed * dt
		}
		if st.MoveRight {
			cam.X += s.speed * dt
		}
		if st.MoveUp {
			cam.Y -= s.speed * dt
		}
		if st.MoveDown {
			cam.Y += s.speed * dt
		}
	}

	cam.X = clamp(cam.X, 0, maxScroll(gameconfig.Global.WorldWidth(), bounds.W))
	cam.Y = clamp(cam.Y, 0, maxScroll(gameconfig.Global.WorldHeight(), bounds.H))
}

// maxScroll returns the largest camera offset that keeps the viewport inside
// the world, or 0 if the world is smaller than the viewport.
func maxScroll(worldSize, viewSize float64) float64 {
	if m := worldSize - viewSize; m > 0 {
		return m
	}
	return 0
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

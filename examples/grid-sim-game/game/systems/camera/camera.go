// Package camera implements CameraScrollSystem, a free-scrolling camera
// driven directly by player input (arrow keys / WASD / gamepad), clamped to
// the grid's world bounds with chrome overscroll.
package camera

import (
	"example.com/grid-sim-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// CameraScrollSystem moves the World's Camera resource continuously while an
// arrow key (or WASD, or gamepad d-pad/stick - all merged into Input.MoveX by
// the engine) is held. Clamping allows a little overscroll past the map edges
// so rows under the toolbar (and the far edge against the side panel) can be
// brought fully into the clear playfield. Unlike systems.CameraFollow (which
// centers on an entity), this is free-scrolling driven by input, so it lives
// here rather than in the engine's generic systems package.
type CameraScrollSystem struct {
	speed float64
}

// NewCameraScrollSystem builds the system with the given scroll speed (pixels
// per second).
func NewCameraScrollSystem(speed float64) *CameraScrollSystem {
	return &CameraScrollSystem{speed: speed}
}

// Update advances the camera and clamps it with chrome-aware overscroll.
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

	minX, maxX, minY, maxY := scrollLimits(bounds.W, bounds.H)
	cam.X = clamp(cam.X, minX, maxX)
	cam.Y = clamp(cam.Y, minY, maxY)
}

// scrollLimits returns inclusive camera offset bounds for the playfield.
//
// Negative minY (-ToolbarHeight) scrolls the top of the map down from under
// the toolbar. maxX uses the left playfield width (not the full screen) so the
// rightmost columns can sit clear of the ImGui panel.
func scrollLimits(screenW, screenH float64) (minX, maxX, minY, maxY float64) {
	cfg := gameconfig.Global
	viewW := cfg.PlayfieldWidth(screenW)

	minX = 0
	maxX = cfg.WorldWidth() - viewW
	if maxX < minX {
		maxX = minX
	}

	minY = -cfg.ToolbarHeight
	maxY = cfg.WorldHeight() - screenH
	if maxY < minY {
		// Map shorter than the screen: pin so the top row clears the toolbar.
		maxY = minY
	}
	return minX, maxX, minY, maxY
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

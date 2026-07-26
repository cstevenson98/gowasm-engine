package systems

import (
	"github.com/cstevenson98/milo/pkg/components"
	"github.com/cstevenson98/milo/pkg/ecs"
)

// CameraFollow centers the World's Camera resource on the entity tagged with
// CameraTarget, every frame. It requires a Camera resource (seeded by
// BaseState.Enter) and reads ScreenBounds to compute the centering offset. With
// no CameraTarget entity present it is a no-op, so adding the system to a
// state's Schedule is harmless even before any entity opts in.
type CameraFollow struct {
	target *ecs.Filter2[components.Position, components.CameraTarget]
}

// NewCameraFollow builds the CameraFollow system for world w.
func NewCameraFollow(w *ecs.World) *CameraFollow {
	return &CameraFollow{
		target: ecs.NewFilter2[components.Position, components.CameraTarget](w),
	}
}

// Update re-centers the camera on the (first) CameraTarget entity found.
func (c *CameraFollow) Update(w *ecs.World, dt float64) {
	cam := ecs.GetResource[components.Camera](w)
	bounds := ecs.GetResource[components.ScreenBounds](w)
	if cam == nil || bounds == nil {
		return
	}

	zoom := cam.Zoom
	if zoom <= 0 {
		zoom = 1
	}

	c.target.Each(func(_ ecs.Entity, p *components.Position, _ *components.CameraTarget) {
		cam.X = p.X - (bounds.W/zoom)/2
		cam.Y = p.Y - (bounds.H/zoom)/2
	})
}

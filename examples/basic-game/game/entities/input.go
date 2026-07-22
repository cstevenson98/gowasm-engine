package entities

import (
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
)

// PlayerInputSystem converts the current input snapshot (the Input resource)
// into velocity on player-controlled entities. Replaces Player.HandleInput.
type PlayerInputSystem struct {
	f *ecs.Filter2[PlayerControl, components.Velocity]
}

// NewPlayerInputSystem builds the system for world w.
func NewPlayerInputSystem(w *ecs.World) *PlayerInputSystem {
	return &PlayerInputSystem{
		f: ecs.NewFilter2[PlayerControl, components.Velocity](w),
	}
}

// Update reads the Input resource and sets each player's velocity, normalizing
// diagonal movement to preserve speed.
func (s *PlayerInputSystem) Update(w *ecs.World, dt float64) {
	in := ecs.GetResource[components.Input](w)
	if in == nil {
		return
	}
	st := in.State

	s.f.Each(func(_ ecs.Entity, pc *PlayerControl, v *components.Velocity) {
		var vx, vy float64
		if st.MoveLeft {
			vx -= pc.Speed
		}
		if st.MoveRight {
			vx += pc.Speed
		}
		if st.MoveUp {
			vy -= pc.Speed
		}
		if st.MoveDown {
			vy += pc.Speed
		}
		if vx != 0 && vy != 0 {
			vx *= 0.7071 // 1/sqrt(2)
			vy *= 0.7071
		}
		v.DX = vx
		v.DY = vy
	})
}

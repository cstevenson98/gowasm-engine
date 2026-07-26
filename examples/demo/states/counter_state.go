package states

import (
	"fmt"

	"github.com/cstevenson98/milo/pkg/state"
	"github.com/cstevenson98/milo/pkg/types"
)

// counterScale is the glyph scale for the centered counter on a 720p screen.
const counterScale = 8

// darkBlue is the demo clear color (deep navy).
var darkBlue = types.Color{0.05, 0.08, 0.22, 1}

// CounterState is a one-state demo: Up arrow increments a counter drawn as
// large centered text.
type CounterState struct {
	*state.BaseState
	n int
}

// NewCounterState constructs the demo state.
func NewCounterState() *CounterState {
	return &CounterState{BaseState: state.NewBaseState("Counter")}
}

// Enter seeds the world via BaseState.
func (s *CounterState) Enter(deps state.Deps) error {
	return s.BaseState.Enter(deps)
}

// Update increments the counter on Up edge and runs the base schedule.
func (s *CounterState) Update(dt float64) {
	in := s.Input()
	if in.UpPressed && !in.UpPressedLastFrame {
		s.n++
	}
	s.BaseState.Update(dt)
}

// DrawOverlays fills a dark blue background, then draws the counter.
func (s *CounterState) DrawOverlays() error {
	ui := s.UI()
	ui.Rect(0, 0, s.ScreenWidth(), s.ScreenHeight(), darkBlue)

	h := ui.LineHeightScaled(counterScale)
	y := (s.ScreenHeight() - h) / 2
	ui.TextCenteredScaled(y, counterScale, types.White, fmt.Sprintf("%d", s.n))
	lh := ui.LineHeight()
	ui.TextCentered(s.ScreenHeight()-lh*3, types.Gray, "Up arrow: +1")
	ui.TextCentered(s.ScreenHeight()-lh*2, types.Gray, "ESC: quit")
	return s.BaseState.DrawOverlays()
}

package ecs

// System contains behaviour that runs against a World each frame. Systems hold
// their filters/mappers as fields (built once in a constructor) and iterate
// them in Update.
type System interface {
	Update(w *World, dt float64)
}

// SystemFunc adapts a plain function to the System interface.
type SystemFunc func(w *World, dt float64)

func (f SystemFunc) Update(w *World, dt float64) { f(w, dt) }

// Schedule runs an ordered list of systems. Order is significant: systems run
// in the order they were added.
type Schedule struct {
	systems []System
}

// NewSchedule creates a Schedule from an ordered list of systems.
func NewSchedule(systems ...System) *Schedule {
	return &Schedule{systems: systems}
}

// Add appends a system to the end of the schedule and returns the schedule for
// chaining.
func (s *Schedule) Add(sys System) *Schedule {
	s.systems = append(s.systems, sys)
	return s
}

// Run executes every system once, in order.
func (s *Schedule) Run(w *World, dt float64) {
	for _, sys := range s.systems {
		sys.Update(w, dt)
	}
}

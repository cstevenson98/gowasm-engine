package ecs

import "testing"

type position struct{ X, Y float64 }
type velocity struct{ DX, DY float64 }
type tag struct{}
type screenBounds struct{ W, H float64 }

func TestMapAndFilter2(t *testing.T) {
	w := NewWorld()
	m := NewMap2[position, velocity](w)

	for i := 0; i < 10; i++ {
		m.NewEntity(&position{X: float64(i)}, &velocity{DX: 1})
	}

	f := NewFilter2[position, velocity](w)
	if got := f.Count(); got != 10 {
		t.Fatalf("Count() = %d, want 10", got)
	}

	f.Each(func(_ Entity, p *position, v *velocity) {
		p.X += v.DX
	})

	sum := 0.0
	f.Each(func(_ Entity, p *position, _ *velocity) { sum += p.X })
	// original sum 0..9 = 45, plus 10 increments of 1 = 55
	if sum != 55 {
		t.Fatalf("sum after move = %v, want 55", sum)
	}
}

func TestFilterWithWithout(t *testing.T) {
	w := NewWorld()
	withTag := NewMap2[position, tag](w)
	noTag := NewMap1[position](w)

	withTag.NewEntity(&position{}, &tag{})
	withTag.NewEntity(&position{}, &tag{})
	noTag.NewEntity(&position{})

	if got := NewFilter1[position](w).With(C[tag]()).Count(); got != 2 {
		t.Fatalf("With(tag) count = %d, want 2", got)
	}
	if got := NewFilter1[position](w).Without(C[tag]()).Count(); got != 1 {
		t.Fatalf("Without(tag) count = %d, want 1", got)
	}
}

func TestEntityLifecycle(t *testing.T) {
	w := NewWorld()
	m := NewMap1[position](w)
	e := m.NewEntity(&position{X: 5})

	if !w.Alive(e) {
		t.Fatal("entity should be alive")
	}
	if !m.Has(e) {
		t.Fatal("entity should have position")
	}
	if p := m.Get(e); p.X != 5 {
		t.Fatalf("Get().X = %v, want 5", p.X)
	}
	w.Remove(e)
	if w.Alive(e) {
		t.Fatal("entity should be dead after Remove")
	}
}

func TestResources(t *testing.T) {
	w := NewWorld()
	if HasResource[screenBounds](w) {
		t.Fatal("resource should be absent initially")
	}
	if GetResource[screenBounds](w) != nil {
		t.Fatal("GetResource should be nil when absent")
	}

	SetResource(w, &screenBounds{W: 800, H: 600})
	if !HasResource[screenBounds](w) {
		t.Fatal("resource should be present")
	}
	if b := GetResource[screenBounds](w); b == nil || b.W != 800 || b.H != 600 {
		t.Fatalf("GetResource = %+v, want 800x600", b)
	}

	SetResource(w, &screenBounds{W: 1024, H: 768})
	if b := GetResource[screenBounds](w); b.W != 1024 {
		t.Fatalf("resource not replaced: W = %v, want 1024", b.W)
	}

	RemoveResource[screenBounds](w)
	if HasResource[screenBounds](w) {
		t.Fatal("resource should be removed")
	}
}

func TestSchedule(t *testing.T) {
	w := NewWorld()
	NewMap2[position, velocity](w).NewEntity(&position{}, &velocity{DX: 2, DY: 3})

	move := &moveSystem{f: NewFilter2[position, velocity](w)}
	order := []string{}
	sched := NewSchedule(
		SystemFunc(func(_ *World, _ float64) { order = append(order, "a") }),
		move,
		SystemFunc(func(_ *World, _ float64) { order = append(order, "b") }),
	)

	sched.Run(w, 1.0)

	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("system order = %v, want [a b]", order)
	}
	move.f.Each(func(_ Entity, p *position, _ *velocity) {
		if p.X != 2 || p.Y != 3 {
			t.Fatalf("move system did not run: pos = %+v", p)
		}
	})
}

type moveSystem struct{ f *Filter2[position, velocity] }

func (s *moveSystem) Update(w *World, dt float64) {
	s.f.Each(func(_ Entity, p *position, v *velocity) {
		p.X += v.DX * dt
		p.Y += v.DY * dt
	})
}

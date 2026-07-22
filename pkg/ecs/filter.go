package ecs

import arkecs "github.com/mlange-42/ark/ecs"

// Filters iterate entities matching a component set. Create them once and
// reuse; call Each once per frame to iterate. Component pointers passed to the
// callback are only valid for the duration of that call.

// Filter1 matches entities with component A.
type Filter1[A any] struct{ inner *arkecs.Filter1[A] }

func NewFilter1[A any](w *World) *Filter1[A] {
	return &Filter1[A]{inner: arkecs.NewFilter1[A](w.inner)}
}

func (f *Filter1[A]) With(comps ...Comp) *Filter1[A] {
	f.inner.With(toArkComps(comps)...)
	return f
}
func (f *Filter1[A]) Without(comps ...Comp) *Filter1[A] {
	f.inner.Without(toArkComps(comps)...)
	return f
}
func (f *Filter1[A]) Exclusive() *Filter1[A] { f.inner.Exclusive(); return f }

func (f *Filter1[A]) Each(fn func(e Entity, a *A)) {
	q := f.inner.Query()
	for q.Next() {
		fn(Entity(q.Entity()), q.Get())
	}
}

func (f *Filter1[A]) Count() int {
	q := f.inner.Query()
	n := q.Count()
	q.Close()
	return n
}

// Filter2 matches entities with components A and B.
type Filter2[A, B any] struct{ inner *arkecs.Filter2[A, B] }

func NewFilter2[A, B any](w *World) *Filter2[A, B] {
	return &Filter2[A, B]{inner: arkecs.NewFilter2[A, B](w.inner)}
}

func (f *Filter2[A, B]) With(comps ...Comp) *Filter2[A, B] {
	f.inner.With(toArkComps(comps)...)
	return f
}
func (f *Filter2[A, B]) Without(comps ...Comp) *Filter2[A, B] {
	f.inner.Without(toArkComps(comps)...)
	return f
}
func (f *Filter2[A, B]) Exclusive() *Filter2[A, B] { f.inner.Exclusive(); return f }

func (f *Filter2[A, B]) Each(fn func(e Entity, a *A, b *B)) {
	q := f.inner.Query()
	for q.Next() {
		a, b := q.Get()
		fn(Entity(q.Entity()), a, b)
	}
}

func (f *Filter2[A, B]) Count() int {
	q := f.inner.Query()
	n := q.Count()
	q.Close()
	return n
}

// Filter3 matches entities with components A, B and C.
type Filter3[A, B, C any] struct{ inner *arkecs.Filter3[A, B, C] }

func NewFilter3[A, B, C any](w *World) *Filter3[A, B, C] {
	return &Filter3[A, B, C]{inner: arkecs.NewFilter3[A, B, C](w.inner)}
}

func (f *Filter3[A, B, C]) With(comps ...Comp) *Filter3[A, B, C] {
	f.inner.With(toArkComps(comps)...)
	return f
}
func (f *Filter3[A, B, C]) Without(comps ...Comp) *Filter3[A, B, C] {
	f.inner.Without(toArkComps(comps)...)
	return f
}
func (f *Filter3[A, B, C]) Exclusive() *Filter3[A, B, C] { f.inner.Exclusive(); return f }

func (f *Filter3[A, B, C]) Each(fn func(e Entity, a *A, b *B, c *C)) {
	q := f.inner.Query()
	for q.Next() {
		a, b, c := q.Get()
		fn(Entity(q.Entity()), a, b, c)
	}
}

func (f *Filter3[A, B, C]) Count() int {
	q := f.inner.Query()
	n := q.Count()
	q.Close()
	return n
}

// Filter4 matches entities with components A, B, C and D.
type Filter4[A, B, C, D any] struct{ inner *arkecs.Filter4[A, B, C, D] }

func NewFilter4[A, B, C, D any](w *World) *Filter4[A, B, C, D] {
	return &Filter4[A, B, C, D]{inner: arkecs.NewFilter4[A, B, C, D](w.inner)}
}

func (f *Filter4[A, B, C, D]) With(comps ...Comp) *Filter4[A, B, C, D] {
	f.inner.With(toArkComps(comps)...)
	return f
}
func (f *Filter4[A, B, C, D]) Without(comps ...Comp) *Filter4[A, B, C, D] {
	f.inner.Without(toArkComps(comps)...)
	return f
}
func (f *Filter4[A, B, C, D]) Exclusive() *Filter4[A, B, C, D] { f.inner.Exclusive(); return f }

func (f *Filter4[A, B, C, D]) Each(fn func(e Entity, a *A, b *B, c *C, d *D)) {
	q := f.inner.Query()
	for q.Next() {
		a, b, c, d := q.Get()
		fn(Entity(q.Entity()), a, b, c, d)
	}
}

func (f *Filter4[A, B, C, D]) Count() int {
	q := f.inner.Query()
	n := q.Count()
	q.Close()
	return n
}

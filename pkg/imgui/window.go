package imgui

import "fmt"

// WindowBuilder is a fluent, library-hiding DSL for populating an ImGui window.
// Methods are nil-safe and no-op when ImGui is inactive (WASM stubs / nil ctx).
type WindowBuilder struct {
	ctx *Context
}

// Text draws a formatted text line.
func (w *WindowBuilder) Text(format string, args ...any) {
	if w == nil || w.ctx == nil || !w.ctx.ready {
		return
	}
	w.ctx.textPlatform(fmt.Sprintf(format, args...))
}

// Separator draws a horizontal rule.
func (w *WindowBuilder) Separator() {
	if w == nil || w.ctx == nil || !w.ctx.ready {
		return
	}
	w.ctx.separatorPlatform()
}

// Checkbox draws a checkbox bound to v. Returns true when the value changed.
func (w *WindowBuilder) Checkbox(label string, v *bool) bool {
	if w == nil || w.ctx == nil || !w.ctx.ready || v == nil {
		return false
	}
	return w.ctx.checkboxPlatform(label, v)
}

// SliderFloat draws a float slider bound to v (min..max). Returns true when
// the value changed. v is float64 for convenience; values are converted to
// float32 for ImGui.
func (w *WindowBuilder) SliderFloat(label string, v *float64, min, max float64) bool {
	if w == nil || w.ctx == nil || !w.ctx.ready || v == nil {
		return false
	}
	return w.ctx.sliderFloatPlatform(label, v, min, max)
}

// Button draws a button. If onClick is non-nil it is invoked when the button
// is pressed this frame. Returns true when pressed.
func (w *WindowBuilder) Button(label string, onClick func()) bool {
	if w == nil || w.ctx == nil || !w.ctx.ready {
		return false
	}
	pressed := w.ctx.buttonPlatform(label)
	if pressed && onClick != nil {
		onClick()
	}
	return pressed
}

// SameLine places the next widget on the same line as the previous one.
func (w *WindowBuilder) SameLine() {
	if w == nil || w.ctx == nil || !w.ctx.ready {
		return
	}
	w.ctx.sameLinePlatform()
}

// TreeNode draws a collapsible tree node and runs fn while it is open.
func (w *WindowBuilder) TreeNode(label string, fn func(w *WindowBuilder)) {
	if w == nil || w.ctx == nil || !w.ctx.ready || fn == nil {
		return
	}
	w.ctx.treeNodePlatform(label, fn)
}

// Columns starts (or changes) an ImGui multi-column layout with count columns.
// Pass count=1 to return to a single column. Pair with NextColumn to advance.
func (w *WindowBuilder) Columns(count int) {
	if w == nil || w.ctx == nil || !w.ctx.ready || count < 1 {
		return
	}
	w.ctx.columnsPlatform(count)
}

// NextColumn advances to the next column in the current Columns layout.
func (w *WindowBuilder) NextColumn() {
	if w == nil || w.ctx == nil || !w.ctx.ready {
		return
	}
	w.ctx.nextColumnPlatform()
}

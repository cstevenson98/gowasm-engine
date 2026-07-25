// Package imgui provides an opt-in Dear ImGui facade for the engine.
//
// Call engine.EnableImGui() before Initialize to activate it. States that want
// to draw ImGui windows implement StateRenderer; the engine invokes
// RenderImGui during Update (between NewFrame and EndFrame), then blits in Draw.
//
// On desktop this package wraps AllenDang/cimgui-go's ebiten backend. On
// WebAssembly (GOOS=js) every method is a silent no-op so games still compile
// and run without CGo.
//
// Only this package imports the ImGui library (and ImPlot); game code talks
// exclusively to Context / WindowBuilder / PlotBuilder.
package imgui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// StateRenderer is an optional interface a state.State may implement to draw
// ImGui windows. The engine calls RenderImGui on the Update path (after
// NewFrame, before EndFrame) when ImGui is enabled — not from Draw, because
// Ebiten may run Draw on another thread and ImGui requires a single-threaded
// frame scope.
type StateRenderer interface {
	RenderImGui(ctx *Context)
}

// Context owns the ImGui backend for one engine instance. All methods are
// nil-safe: a nil *Context (or a WASM stub) is a silent no-op.
type Context struct {
	screenW, screenH int
	ready            bool
	be               backendHandle
}

// NewContext allocates an uninitialized ImGui context. Call Init before use.
func NewContext() *Context {
	return &Context{}
}

// Init creates the underlying ImGui backend using the given virtual screen
// size. Safe to call more than once; subsequent calls are no-ops.
func (c *Context) Init(screenW, screenH int) error {
	if c == nil || c.ready {
		return nil
	}
	c.screenW = screenW
	c.screenH = screenH
	return c.initPlatform()
}

// SetScreenSize updates the display size ImGui uses for layout (e.g. after a
// virtual-resolution change).
func (c *Context) SetScreenSize(screenW, screenH int) {
	if c == nil {
		return
	}
	c.screenW = screenW
	c.screenH = screenH
}

// NewFrame syncs input and starts a new ImGui frame. Call once per Update.
func (c *Context) NewFrame() {
	if c == nil || !c.ready {
		return
	}
	c.newFramePlatform()
}

// EndFrame finishes the ImGui frame after widgets have been declared.
func (c *Context) EndFrame() {
	if c == nil || !c.ready {
		return
	}
	c.endFramePlatform()
}

// Draw renders the finished ImGui frame onto screen. Call from Engine.Draw
// after EndFrame.
func (c *Context) Draw(screen *ebiten.Image) {
	if c == nil || !c.ready || screen == nil {
		return
	}
	c.drawPlatform(screen)
}

// Window opens a named ImGui window and runs fn to populate it. Begin/End are
// handled automatically. fn is skipped when the window is collapsed/closed or
// when ImGui is inactive.
func (c *Context) Window(title string, fn func(w *WindowBuilder)) {
	if c == nil || !c.ready || fn == nil {
		return
	}
	c.windowPlatform(title, fn)
}

// Panel opens a fixed-position, non-movable ImGui window covering the given
// rectangle (in virtual screen pixels). Useful for side docks / half-screen
// inspector panels.
func (c *Context) Panel(title string, x, y, w, h float64, fn func(wb *WindowBuilder)) {
	if c == nil || !c.ready || fn == nil {
		return
	}
	c.panelPlatform(title, x, y, w, h, fn)
}

// Ready reports whether Init succeeded and ImGui is usable.
func (c *Context) Ready() bool {
	return c != nil && c.ready
}

//go:build js

package imgui

import "github.com/hajimehoshi/ebiten/v2"

// backendHandle is unused on WASM; ImGui is a silent no-op (no CGo).
type backendHandle struct{}

func (c *Context) initPlatform() error {
	// Keep ready=false so all DSL methods no-op. Returning nil lets EnableImGui
	// succeed on WASM without pulling in cimgui-go.
	return nil
}

func (c *Context) newFramePlatform() {}

func (c *Context) endFramePlatform() {}

func (c *Context) drawPlatform(screen *ebiten.Image) {}

func (c *Context) windowPlatform(title string, fn func(w *WindowBuilder)) {}

func (c *Context) panelPlatform(title string, x, y, w, h float64, fn func(wb *WindowBuilder)) {
}

func (c *Context) textPlatform(s string) {}

func (c *Context) separatorPlatform() {}

func (c *Context) checkboxPlatform(label string, v *bool) bool { return false }

func (c *Context) sliderFloatPlatform(label string, v *float64, min, max float64) bool {
	return false
}

func (c *Context) buttonPlatform(label string) bool { return false }

func (c *Context) treeNodePlatform(label string, fn func(w *WindowBuilder)) {}

func (c *Context) columnsPlatform(count int) {}

func (c *Context) nextColumnPlatform() {}

func (c *Context) plotPlatform(title string, height float64, fn func(p *PlotBuilder)) {}

func (c *Context) plotSetupAxesPlatform(xLabel, yLabel string) {}

func (c *Context) plotLinePlatform(label string, ys []float64) {}

func (c *Context) plotLineXYPlatform(label string, xs, ys []float64) {}

func (c *Context) plotBarsPlatform(label string, ys []float64) {}

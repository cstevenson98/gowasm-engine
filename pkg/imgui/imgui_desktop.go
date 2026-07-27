//go:build !js

package imgui

import (
	ebitenbackend "github.com/AllenDang/cimgui-go/backend/ebiten-backend"
	cim "github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/AllenDang/cimgui-go/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

// backendHandle is the desktop ImGui ebiten backend.
type backendHandle = *ebitenbackend.EbitenBackend

const defaultPlotHeight = 180.0

func (c *Context) initPlatform() error {
	be := ebitenbackend.NewEbitenBackend()
	// Transparent clear so ImGui overlays the game instead of wiping it.
	be.SetBgColor(cim.NewVec4(0, 0, 0, 0))
	be.CreateWindow("milo-imgui", c.screenW, c.screenH)
	// CreateWindow also sets the OS window size/title; Layout syncs display size.
	be.Layout(c.screenW, c.screenH)
	// ImPlot needs its own context, created after the ImGui context exists.
	implot.CreateContext()
	c.be = be
	c.ready = true
	return nil
}

func (c *Context) newFramePlatform() {
	c.be.Layout(c.screenW, c.screenH)
	c.be.BeginFrame()
}

func (c *Context) endFramePlatform() {
	c.be.EndFrame()
}

func (c *Context) drawPlatform(screen *ebiten.Image) {
	c.be.Draw(screen)
}

func (c *Context) windowPlatform(title string, fn func(w *WindowBuilder)) {
	if !cim.Begin(title) {
		cim.End()
		return
	}
	fn(&WindowBuilder{ctx: c})
	cim.End()
}

func (c *Context) panelPlatform(title string, x, y, w, h float64, fn func(wb *WindowBuilder)) {
	cim.SetNextWindowPos(cim.NewVec2(float32(x), float32(y)))
	cim.SetNextWindowSize(cim.NewVec2(float32(w), float32(h)))
	flags := cim.WindowFlagsNoResize | cim.WindowFlagsNoMove | cim.WindowFlagsNoCollapse
	if !cim.BeginV(title, nil, flags) {
		cim.End()
		return
	}
	fn(&WindowBuilder{ctx: c})
	cim.End()
}

func (c *Context) textPlatform(s string) {
	cim.TextUnformatted(s)
}

func (c *Context) separatorPlatform() {
	cim.Separator()
}

func (c *Context) checkboxPlatform(label string, v *bool) bool {
	return cim.Checkbox(label, v)
}

func (c *Context) sliderFloatPlatform(label string, v *float64, min, max float64) bool {
	f := float32(*v)
	changed := cim.SliderFloat(label, &f, float32(min), float32(max))
	if changed {
		*v = float64(f)
	}
	return changed
}

func (c *Context) buttonPlatform(label string) bool {
	return cim.Button(label)
}

func (c *Context) sameLinePlatform() {
	cim.SameLine()
}

func (c *Context) treeNodePlatform(label string, fn func(w *WindowBuilder)) {
	if !cim.TreeNodeStr(label) {
		return
	}
	fn(&WindowBuilder{ctx: c})
	cim.TreePop()
}

func (c *Context) columnsPlatform(count int) {
	cim.ColumnsV(int32(count), "", true)
}

func (c *Context) nextColumnPlatform() {
	cim.NextColumn()
}

func (c *Context) plotPlatform(title string, height float64, fn func(p *PlotBuilder)) {
	if height <= 0 {
		height = defaultPlotHeight
	}
	// Axis fit is controlled per-plot via SetupAxes / SetupAxesYLimits
	// (AxisFlagsAutoFit / SetupAxisLimits), not SetNextAxesToFit, so fixed Y
	// limits are not overwritten.
	if !implot.BeginPlotV(title, cim.NewVec2(-1, float32(height)), 0) {
		return
	}
	fn(&PlotBuilder{ctx: c})
	implot.EndPlot()
}

func (c *Context) plotSetupAxesPlatform(xLabel, yLabel string) {
	// AutoFit keeps axes locked to the data every frame (not only the first).
	implot.SetupAxesV(xLabel, yLabel, implot.AxisFlagsAutoFit, implot.AxisFlagsAutoFit)
}

func (c *Context) plotSetupAxesYLimitsPlatform(xLabel, yLabel string, yMin, yMax float64) {
	implot.SetupAxesV(xLabel, yLabel, implot.AxisFlagsAutoFit, 0)
	implot.SetupAxisLimitsV(implot.AxisY1, yMin, yMax, implot.CondAlways)
}

func (c *Context) plotSetupAxesXLimitsPlatform(xLabel, yLabel string, xMin, xMax float64) {
	implot.SetupAxesV(xLabel, yLabel, 0, implot.AxisFlagsAutoFit)
	implot.SetupAxisLimitsV(implot.AxisX1, xMin, xMax, implot.CondAlways)
}

func (c *Context) plotLinePlatform(label string, ys []float64) {
	implot.PlotLinedoublePtrInt(label, utils.SliceToPtr(ys), int32(len(ys)))
}

func (c *Context) plotLineXYPlatform(label string, xs, ys []float64) {
	implot.PlotLinedoublePtrdoublePtr(label, utils.SliceToPtr(xs), utils.SliceToPtr(ys), int32(len(xs)))
}

func (c *Context) plotBarsPlatform(label string, ys []float64) {
	implot.PlotBarsdoublePtrInt(label, utils.SliceToPtr(ys), int32(len(ys)))
}

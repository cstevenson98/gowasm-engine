package imgui

// PlotBuilder populates an ImPlot region opened by WindowBuilder.Plot.
// Methods are nil-safe and no-op when ImGui/ImPlot is inactive.
type PlotBuilder struct {
	ctx *Context
}

// Plot opens an ImPlot chart of the given height in pixels (0 uses a default)
// and runs fn to add series. BeginPlot/EndPlot are handled automatically.
func (w *WindowBuilder) Plot(title string, height float64, fn func(p *PlotBuilder)) {
	if w == nil || w.ctx == nil || !w.ctx.ready || fn == nil {
		return
	}
	w.ctx.plotPlatform(title, height, fn)
}

// SetupAxes sets optional X/Y axis labels for the current plot. Both axes use
// ImPlot AutoFit so the view recenters every frame as series data changes.
func (p *PlotBuilder) SetupAxes(xLabel, yLabel string) {
	if p == nil || p.ctx == nil || !p.ctx.ready {
		return
	}
	p.ctx.plotSetupAxesPlatform(xLabel, yLabel)
}

// SetupAxesYLimits labels the axes, AutoFits X to the series, and locks Y to
// [yMin, yMax] every frame (no Y AutoFit).
func (p *PlotBuilder) SetupAxesYLimits(xLabel, yLabel string, yMin, yMax float64) {
	if p == nil || p.ctx == nil || !p.ctx.ready {
		return
	}
	p.ctx.plotSetupAxesYLimitsPlatform(xLabel, yLabel, yMin, yMax)
}

// SetupAxesXLimits labels the axes, locks X to [xMin, xMax] every frame, and
// AutoFits Y to the series (no X AutoFit).
func (p *PlotBuilder) SetupAxesXLimits(xLabel, yLabel string, xMin, xMax float64) {
	if p == nil || p.ctx == nil || !p.ctx.ready {
		return
	}
	p.ctx.plotSetupAxesXLimitsPlatform(xLabel, yLabel, xMin, xMax)
}

// Line plots ys against index 0..n-1.
func (p *PlotBuilder) Line(label string, ys []float64) {
	if p == nil || p.ctx == nil || !p.ctx.ready || len(ys) == 0 {
		return
	}
	p.ctx.plotLinePlatform(label, ys)
}

// LineXY plots (xs[i], ys[i]). Lengths must match; the shorter length wins.
func (p *PlotBuilder) LineXY(label string, xs, ys []float64) {
	if p == nil || p.ctx == nil || !p.ctx.ready || len(xs) == 0 || len(ys) == 0 {
		return
	}
	n := len(xs)
	if len(ys) < n {
		n = len(ys)
	}
	p.ctx.plotLineXYPlatform(label, xs[:n], ys[:n])
}

// Bars plots ys as a bar series against index 0..n-1.
func (p *PlotBuilder) Bars(label string, ys []float64) {
	if p == nil || p.ctx == nil || !p.ctx.ready || len(ys) == 0 {
		return
	}
	p.ctx.plotBarsPlatform(label, ys)
}

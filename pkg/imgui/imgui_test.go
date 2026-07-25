package imgui

import "testing"

func TestNilContextSafe(t *testing.T) {
	var c *Context
	c.NewFrame()
	c.Window("x", func(w *WindowBuilder) {
		w.Text("hi %d", 1)
		w.Separator()
		w.Checkbox("on", nil)
		w.SliderFloat("v", nil, 0, 1)
		w.Button("b", nil)
		w.TreeNode("n", func(w *WindowBuilder) { w.Text("nested") })
		w.Columns(2)
		w.NextColumn()
		w.Columns(1)
		w.Plot("p", 100, func(p *PlotBuilder) {
			p.SetupAxes("x", "y")
			p.Line("l", []float64{1, 2})
			p.Bars("b", []float64{1, 2})
			p.LineXY("xy", []float64{0, 1}, []float64{1, 2})
		})
	})
	c.EndFrame()
	c.Draw(nil)
}

func TestUninitializedContextSafe(t *testing.T) {
	c := NewContext()
	if c.Ready() {
		t.Fatal("expected not ready before Init")
	}
	c.NewFrame()
	c.Window("x", func(w *WindowBuilder) { w.Text("should not run") })
	c.EndFrame()
}

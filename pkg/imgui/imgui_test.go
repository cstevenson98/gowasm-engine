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

package entities

import (
	"testing"

	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// newTestWorld returns a World with the resources PlacementSystem/
// CameraScrollSystem expect (mirroring what GridState.Enter/BaseState.Enter
// seed), so system tests don't need a full state.
func newTestWorld() *ecs.World {
	w := ecs.NewWorld()
	ecs.SetResource(w, &components.ScreenBounds{W: 320, H: 240})
	ecs.SetResource(w, &components.Input{})
	ecs.SetResource(w, &components.Camera{Zoom: 1})
	ecs.SetResource(w, &PlacementState{Tool: ToolNone})
	ecs.SetResource(w, NewGridOccupancy())
	return w
}

func setMouseClick(w *ecs.World, x, y float64) {
	ecs.SetResource(w, &components.Input{State: types.InputState{
		Mouse: types.MouseState{X: x, Y: y, Left: types.MouseButtonState{Pressed: true}},
	}})
}

func TestManhattanPathSameRow(t *testing.T) {
	path := ManhattanPath(GridCoord{Col: 2, Row: 3}, GridCoord{Col: 5, Row: 3})
	want := []GridCoord{{2, 3}, {3, 3}, {4, 3}, {5, 3}}
	assertPath(t, path, want)
}

func TestManhattanPathSameColumn(t *testing.T) {
	path := ManhattanPath(GridCoord{Col: 2, Row: 3}, GridCoord{Col: 2, Row: 7})
	want := []GridCoord{{2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}}
	assertPath(t, path, want)
}

func TestManhattanPathLShape(t *testing.T) {
	path := ManhattanPath(GridCoord{Col: 2, Row: 3}, GridCoord{Col: 5, Row: 7})
	want := []GridCoord{{2, 3}, {3, 3}, {4, 3}, {5, 3}, {5, 4}, {5, 5}, {5, 6}, {5, 7}}
	assertPath(t, path, want)
}

func TestManhattanPathSameCell(t *testing.T) {
	path := ManhattanPath(GridCoord{Col: 4, Row: 4}, GridCoord{Col: 4, Row: 4})
	assertPath(t, path, []GridCoord{{4, 4}})
}

func assertPath(t *testing.T, got, want []GridCoord) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("path = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("path[%d] = %v, want %v (full: got=%v want=%v)", i, got[i], want[i], got, want)
		}
	}
}

func TestSpawnGeneratorComponents(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnGenerator(w, GridCoord{Col: 1, Row: 2})

	obj := ecs.NewMap1[GridObject](w).Get(e)
	if obj == nil || obj.Kind != ToolGenerator || obj.Cell != (GridCoord{Col: 1, Row: 2}) {
		t.Fatalf("GridObject = %+v, want Kind=Generator Cell=(1,2)", obj)
	}
	pos := ecs.NewMap1[components.Position](w).Get(e)
	ts := 32.0 // gameconfig.Global.TileSize at time of writing
	if pos.X != ts*1 || pos.Y != ts*2 {
		t.Fatalf("Position = %+v, want (%v, %v)", pos, ts, ts*2)
	}
	if !ecs.NewMap1[components.LayerEntities](w).Has(e) {
		t.Fatal("generator should be on the ENTITIES layer")
	}
}

func TestSpawnBlankIsBackground(t *testing.T) {
	w := ecs.NewWorld()
	e := SpawnBlank(w, GridCoord{Col: 0, Row: 0})
	if !ecs.NewMap1[components.LayerBackground](w).Has(e) {
		t.Fatal("blank tile should be on the BACKGROUND layer")
	}
}

func TestGridOccupancy(t *testing.T) {
	occ := NewGridOccupancy()
	cell := GridCoord{Col: 1, Row: 1}
	if occ.Occupied(cell) {
		t.Fatal("fresh occupancy should report free")
	}
	occ.Occupy(cell, ecs.Entity{})
	if !occ.Occupied(cell) {
		t.Fatal("cell should be occupied after Occupy")
	}
}

func TestPlacementSystemSelectAndPlaceGenerator(t *testing.T) {
	w := newTestWorld()
	sys := NewPlacementSystem(w)

	// Click the GEN button in the toolbar.
	genBtn := findButton(t, ToolGenerator)
	setMouseClick(w, genBtn.X+1, genBtn.Y+1)
	sys.Update(w, 0)

	placement := ecs.GetResource[PlacementState](w)
	if placement.Tool != ToolGenerator {
		t.Fatalf("Tool = %v, want ToolGenerator", placement.Tool)
	}

	// Click a grid cell below the toolbar; must actually place a generator.
	setMouseClick(w, 100, 100) // world (100,100) at cam (0,0) -> cell (3,3) at TileSize 32
	sys.Update(w, 0)

	occ := ecs.GetResource[GridOccupancy](w)
	cell := GridCoord{Col: 3, Row: 3}
	if !occ.Occupied(cell) {
		t.Fatal("expected cell (3,3) to be occupied after placing a generator")
	}

	count := 0
	ecs.NewFilter1[GridObject](w).Each(func(_ ecs.Entity, o *GridObject) {
		if o.Kind == ToolGenerator {
			count++
		}
	})
	if count != 1 {
		t.Fatalf("expected exactly 1 generator entity, got %d", count)
	}
}

func TestPlacementSystemTogglesToolOff(t *testing.T) {
	w := newTestWorld()
	sys := NewPlacementSystem(w)
	btn := findButton(t, ToolHouse)

	setMouseClick(w, btn.X+1, btn.Y+1)
	sys.Update(w, 0)
	if ecs.GetResource[PlacementState](w).Tool != ToolHouse {
		t.Fatal("expected ToolHouse selected")
	}

	// PressedLastFrame must be reset between clicks, or the second click is
	// not seen as a new edge.
	releaseMouse(w)
	setMouseClick(w, btn.X+1, btn.Y+1)
	sys.Update(w, 0)
	if got := ecs.GetResource[PlacementState](w).Tool; got != ToolNone {
		t.Fatalf("clicking the selected tool again should deselect it, got %v", got)
	}
}

func TestPlacementSystemLineTwoClicks(t *testing.T) {
	w := newTestWorld()
	sys := NewPlacementSystem(w)

	lineBtn := findButton(t, ToolLine)
	setMouseClick(w, lineBtn.X+1, lineBtn.Y+1)
	sys.Update(w, 0)
	releaseMouse(w)

	// Start cell (1,1) -> world (32,32).
	setMouseClick(w, 32, 32)
	sys.Update(w, 0)
	if !ecs.GetResource[PlacementState](w).LinePending {
		t.Fatal("expected LinePending after first line click")
	}
	releaseMouse(w)

	// End cell (4,1) -> world (128,32); same row, so path is 4 cells wide.
	setMouseClick(w, 128, 32)
	sys.Update(w, 0)

	placement := ecs.GetResource[PlacementState](w)
	if placement.LinePending {
		t.Fatal("expected LinePending cleared after second line click")
	}

	occ := ecs.GetResource[GridOccupancy](w)
	for col := 1; col <= 4; col++ {
		if !occ.Occupied(GridCoord{Col: col, Row: 1}) {
			t.Fatalf("expected cell (%d,1) occupied by the line", col)
		}
	}
}

func findButton(t *testing.T, tool Tool) ToolbarButton {
	t.Helper()
	for _, b := range ToolbarButtons() {
		if b.Tool == tool {
			return b
		}
	}
	t.Fatalf("no toolbar button for tool %v", tool)
	return ToolbarButton{}
}

func releaseMouse(w *ecs.World) {
	in := ecs.GetResource[components.Input](w)
	st := in.State
	st.Mouse.Left.PressedLastFrame = st.Mouse.Left.Pressed
	st.Mouse.Left.Pressed = false
	ecs.SetResource(w, &components.Input{State: st})
}

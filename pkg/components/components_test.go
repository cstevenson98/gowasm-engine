package components

import "testing"

func TestSpriteTotalFrames(t *testing.T) {
	cases := []struct {
		cols, rows, want int
	}{
		{2, 3, 6},
		{1, 1, 1},
		{0, 0, 1}, // degenerate -> clamped to 1
	}
	for _, c := range cases {
		s := Sprite{Columns: c.cols, Rows: c.rows}
		if got := s.TotalFrames(); got != c.want {
			t.Errorf("TotalFrames(%dx%d) = %d, want %d", c.cols, c.rows, got, c.want)
		}
	}
}

func TestSpriteUV(t *testing.T) {
	s := Sprite{Columns: 2, Rows: 3} // 6 frames, frame w=0.5 h=1/3

	// frame 0 -> top-left
	uv := s.UV()
	if uv.U != 0 || uv.V != 0 || uv.W != 0.5 {
		t.Fatalf("frame 0 UV = %+v", uv)
	}

	// frame 1 -> second column, first row
	s.Frame = 1
	uv = s.UV()
	if uv.U != 0.5 || uv.V != 0 {
		t.Fatalf("frame 1 UV = %+v", uv)
	}

	// frame 2 -> first column, second row
	s.Frame = 2
	uv = s.UV()
	if uv.U != 0 || (uv.V < 0.333 || uv.V > 0.334) {
		t.Fatalf("frame 2 UV = %+v", uv)
	}
}

func TestSpriteUVSingleFrame(t *testing.T) {
	s := Sprite{Columns: 1, Rows: 1}
	uv := s.UV()
	if uv.U != 0 || uv.V != 0 || uv.W != 1 || uv.H != 1 {
		t.Fatalf("single-frame UV = %+v, want full texture", uv)
	}
}

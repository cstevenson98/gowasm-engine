package types

// Color is an RGBA color with components in the range [0, 1].
type Color = [4]float32

// Common UI colors.
var (
	White  = Color{1, 1, 1, 1}
	Black  = Color{0, 0, 0, 1}
	Red    = Color{1, 0, 0, 1}
	Green  = Color{0, 1, 0, 1}
	Blue   = Color{0, 0, 1, 1}
	Yellow = Color{1, 1, 0, 1}
	Gray   = Color{0.7, 0.7, 0.7, 1}
)

// UIManager draws immediate-mode, screen-space UI (text and primitives). The
// engine owns a concrete implementation and injects it into scenes, so scenes
// depend only on this interface rather than a specific UI package.
//
// Methods draw straight to the current screen, so they should be called during
// the render phase (a scene's RenderOverlays), not from Update.
type UIManager interface {
	// Text draws a left-aligned white string with its top-left at (x, y).
	Text(x, y float64, s string)
	// TextColored draws a left-aligned string in the given color at (x, y).
	TextColored(x, y float64, c Color, s string)
	// TextCentered draws a string horizontally centered on screen at height y.
	TextCentered(y float64, c Color, s string)
	// Rect draws a filled rectangle of size (w, h) with its top-left at (x, y).
	Rect(x, y, w, h float64, c Color)
	// Measure returns the rendered width (virtual pixels) of a single-line string.
	Measure(s string) float64
	// LineHeight returns the vertical distance between successive text lines.
	LineHeight() float64
	// ScreenSize returns the virtual screen dimensions the UI centers against.
	ScreenSize() (width, height float64)
}

// NopUI is a UIManager that draws nothing. It is used before a real UI is
// injected (or when UI initialization fails), so UI calls are always safe.
var NopUI UIManager = nopUIManager{}

type nopUIManager struct{}

func (nopUIManager) Text(x, y float64, s string)                 {}
func (nopUIManager) TextColored(x, y float64, c Color, s string) {}
func (nopUIManager) TextCentered(y float64, c Color, s string)   {}
func (nopUIManager) Rect(x, y, w, h float64, c Color)            {}
func (nopUIManager) Measure(s string) float64                    { return 0 }
func (nopUIManager) LineHeight() float64                         { return 0 }
func (nopUIManager) ScreenSize() (float64, float64)              { return 0, 0 }

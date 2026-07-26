package debug

import (
	"fmt"
	"sync"

	"github.com/cstevenson98/milo/pkg/config"
	"github.com/cstevenson98/milo/pkg/types"
)

// ConsoleMessage represents a single debug console message
type ConsoleMessage struct {
	Sender  string
	Message string
	Age     float64 // How long the message has been displayed (seconds)
}

// Config holds the debug console's appearance/behaviour and the virtual
// screen width it draws its background against. NewDebugConsole seeds it with
// the engine's stock defaults (config.Default()) so the console works before
// any engine exists (e.g. in unit tests); Engine.Initialize calls Configure
// with the engine's actual config.Settings so a customized config takes
// effect instead of those defaults.
type Config struct {
	Enabled         bool
	MaxMessages     int
	MessageLifetime float64
	ConsoleHeight   float64
	ScreenWidth     float64
	BackgroundColor [4]float32
	TextColor       [4]float32
}

// DebugConsole handles debug message display
type DebugConsole struct {
	messages []ConsoleMessage
	visible  bool
	mutex    sync.RWMutex
	cfg      Config
}

// Console is the global debug console instance
var Console = NewDebugConsole()

// NewDebugConsole creates a new debug console, seeded with the engine's stock
// config defaults (see Config's doc comment).
func NewDebugConsole() *DebugConsole {
	def := config.Default()
	return &DebugConsole{
		messages: make([]ConsoleMessage, 0),
		visible:  false, // Hidden until toggled (key "3")
		cfg:      configFrom(def),
	}
}

// Configure applies cfg to the console, replacing whatever it was seeded or
// last configured with. Safe to call at any time (e.g. once during engine
// setup); takes the same lock as message operations.
func (dc *DebugConsole) Configure(cfg Config) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.cfg = cfg
}

func configFrom(s config.Settings) Config {
	return Config{
		Enabled:         s.Debug.Enabled,
		MaxMessages:     s.Debug.MaxMessages,
		MessageLifetime: s.Debug.MessageLifetime,
		ConsoleHeight:   s.Debug.ConsoleHeight,
		ScreenWidth:     s.Screen.Width,
		BackgroundColor: s.Debug.BackgroundColor,
		TextColor:       s.Debug.TextColor,
	}
}

// PostMessage adds a message to the console
func (dc *DebugConsole) PostMessage(sender string, message string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	// Add new message
	msg := ConsoleMessage{
		Sender:  sender,
		Message: message,
		Age:     0,
	}

	dc.messages = append(dc.messages, msg)

	// Trim old messages if we exceed max
	if dc.cfg.MaxMessages > 0 && len(dc.messages) > dc.cfg.MaxMessages {
		// Remove oldest message
		dc.messages = dc.messages[1:]
	}
}

// ToggleVisibility toggles the console visibility
func (dc *DebugConsole) ToggleVisibility() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.visible = !dc.visible
}

// SetVisibility sets the console visibility explicitly
func (dc *DebugConsole) SetVisibility(visible bool) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.visible = visible
}

// IsVisible returns true if the console is visible
func (dc *DebugConsole) IsVisible() bool {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return dc.visible
}

// Update updates message ages and removes expired messages
func (dc *DebugConsole) Update(deltaTime float64) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	// If message lifetime is 0, messages never expire
	if dc.cfg.MessageLifetime <= 0 {
		return
	}

	// Age all messages and remove expired ones
	activeMessages := make([]ConsoleMessage, 0, len(dc.messages))
	for _, msg := range dc.messages {
		msg.Age += deltaTime
		if msg.Age < dc.cfg.MessageLifetime {
			activeMessages = append(activeMessages, msg)
		}
	}
	dc.messages = activeMessages
}

// Render draws the debug console using the shared UI manager.
func (dc *DebugConsole) Render(ui types.UIManager) error {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	if !dc.visible || !dc.cfg.Enabled {
		return nil
	}

	if len(dc.messages) == 0 {
		return nil
	}

	consoleHeight := dc.cfg.ConsoleHeight
	consoleWidth := dc.cfg.ScreenWidth

	// Semi-transparent background.
	ui.Rect(0, 0, consoleWidth, consoleHeight, dc.cfg.BackgroundColor)

	// Render messages from top to bottom.
	lineHeight := ui.LineHeight()
	textColor := dc.cfg.TextColor
	y := 5.0 // Top padding

	for _, msg := range dc.messages {
		// Stop if we run out of vertical space.
		if y+lineHeight > consoleHeight {
			break
		}
		ui.TextColored(5, y, textColor, fmt.Sprintf("[%s] %s", msg.Sender, msg.Message))
		y += lineHeight
	}

	return nil
}

// Clear removes all messages from the console
func (dc *DebugConsole) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.messages = make([]ConsoleMessage, 0)
}

// GetMessageCount returns the current number of messages
func (dc *DebugConsole) GetMessageCount() int {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	return len(dc.messages)
}

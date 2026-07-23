package debug

import (
	"fmt"
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// ConsoleMessage represents a single debug console message
type ConsoleMessage struct {
	Sender  string
	Message string
	Age     float64 // How long the message has been displayed (seconds)
}

// DebugConsole handles debug message display
type DebugConsole struct {
	messages    []ConsoleMessage
	visible     bool
	mutex       sync.RWMutex
	maxMessages int
}

// Console is the global debug console instance
var Console = NewDebugConsole()

// NewDebugConsole creates a new debug console
func NewDebugConsole() *DebugConsole {
	return &DebugConsole{
		messages:    make([]ConsoleMessage, 0),
		visible:     false, // Hidden until toggled (key "3")
		maxMessages: config.Global.Debug.MaxMessages,
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
	if len(dc.messages) > dc.maxMessages {
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
	if config.Global.Debug.MessageLifetime <= 0 {
		return
	}

	// Age all messages and remove expired ones
	activeMessages := make([]ConsoleMessage, 0, len(dc.messages))
	for _, msg := range dc.messages {
		msg.Age += deltaTime
		if msg.Age < config.Global.Debug.MessageLifetime {
			activeMessages = append(activeMessages, msg)
		}
	}
	dc.messages = activeMessages
}

// Render draws the debug console using the shared UI manager.
func (dc *DebugConsole) Render(ui types.UIManager) error {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	if !dc.visible || !config.Global.Debug.Enabled {
		return nil
	}

	if len(dc.messages) == 0 {
		return nil
	}

	consoleHeight := config.Global.Debug.ConsoleHeight
	consoleWidth := config.Global.Screen.Width

	// Semi-transparent background.
	ui.Rect(0, 0, consoleWidth, consoleHeight, config.Global.Debug.BackgroundColor)

	// Render messages from top to bottom.
	lineHeight := ui.LineHeight()
	textColor := config.Global.Debug.TextColor
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

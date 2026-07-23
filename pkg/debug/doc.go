// Package debug provides an in-game debug console for surfacing runtime
// messages on screen.
//
// The console is an overlay that systems and states can post short messages to
// (player position, state changes, warnings) and see rendered directly in the
// running game - handy when a terminal is not visible or when debugging input
// and timing that only manifest at runtime.
//
// # Global instance
//
// Console is a process-wide *DebugConsole that any code can post to directly:
//
//	debug.Console.PostMessage("Alert", "Game saved successfully!")
//
// # Rendering and lifetime
//
// DebugConsole keeps a bounded ring of the most recent messages (capped by
// config.Global.Debug.MaxMessages) that optionally fade with age. It is drawn
// as a semi-transparent panel via the types.UIManager facade. BaseState renders
// it automatically as part of DrawOverlays, and toggles its visibility on key 3,
// so any state embedding state.BaseState gets it for free.
package debug

// Package logger provides the engine's leveled logging facility.
//
// All engine packages log through the shared Logger variable rather than
// calling fmt directly, which gives consistent, prefixed output and a single
// place to control verbosity. Logger implements the ILogger interface, so it
// can be swapped for a custom sink (file, browser console, test buffer) by
// assigning a different implementation.
//
// # Levels
//
// Messages are tagged with a severity and suppressed below the configured
// threshold (SetLogLevel), from most to least verbose:
//
//   - Trace - very detailed, safe to emit inside hot loops.
//   - Debug - lower-level detail, but not from inside per-frame loops.
//   - Info  - high-level lifecycle events (engine start, scene changes).
//   - Warn  - recoverable problems worth attention.
//   - Error - failures.
//
// Each level has a plain (Debug) and a formatted (Debugf) variant. The default
// ConsoleLogger writes to standard output at Debug level.
package logger

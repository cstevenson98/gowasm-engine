package types

import (
	"fmt"
)

// DebugMessagePoster is an optional interface that callbacks can use to post debug messages
// This is defined here to avoid circular dependencies
type DebugMessagePoster interface {
	PostMessage(source, message string)
}

// globalDebugPoster is a global debug message poster that can be set by the debug package
var globalDebugPoster DebugMessagePoster

// SetGlobalDebugPoster sets the global debug message poster
func SetGlobalDebugPoster(poster DebugMessagePoster) {
	globalDebugPoster = poster
}

// PostDebugMessageSimple posts a simple debug message with a source string
func PostDebugMessageSimple(source string, format string, args ...interface{}) {
	if globalDebugPoster != nil {
		message := fmt.Sprintf(format, args...)
		globalDebugPoster.PostMessage(source, message)
	}
}

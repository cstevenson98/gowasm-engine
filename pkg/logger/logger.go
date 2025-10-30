package logger

import "fmt"

/*
Log levels:
- Trace: Very detailed logging, allowed in loops
- Debug: Detailed logging at the lower levels of the application, but not inside loops which may clog console
- Info: General information logging at the top level of the application, i.e. provisioning the engine, etc.
- Warn: Warning messages
- Error: Error messages
*/

type LogLevel int

const (
	LogLevelTrace = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var (
	Logger ILogger = NewConsoleLogger()
)

type ILogger interface {
	SetLogLevel(level LogLevel)

	Trace(message string)
	Tracef(format string, args ...interface{})
	Debug(message string)
	Debugf(format string, args ...interface{})
	Info(message string)
	Infof(format string, args ...interface{})
	Warn(message string)
	Warnf(format string, args ...interface{})
	Error(message string)
	Errorf(format string, args ...interface{})
}

type ConsoleLogger struct {
	level LogLevel
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		level: LogLevelDebug,
	}
}

func (l *ConsoleLogger) SetLogLevel(level LogLevel) {
	l.level = level
}

func (l *ConsoleLogger) Trace(message string) {
	if l.level <= LogLevelTrace {
		fmt.Println("TRACE: " + message)
	}
}

func (l *ConsoleLogger) Tracef(format string, args ...interface{}) {
	if l.level <= LogLevelTrace {
		fmt.Println(fmt.Sprintf("TRACE: "+format, args...))
	}
}

func (l *ConsoleLogger) Debug(message string) {
	if l.level <= LogLevelDebug {
		fmt.Println("DEBUG: " + message)
	}
}

func (l *ConsoleLogger) Debugf(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		fmt.Println(fmt.Sprintf("DEBUG: "+format, args...))
	}
}

func (l *ConsoleLogger) Info(message string) {
	if l.level <= LogLevelInfo {
		fmt.Println("INFO: " + message)
	}
}

func (l *ConsoleLogger) Infof(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		fmt.Println(fmt.Sprintf("INFO: "+format, args...))
	}
}

func (l *ConsoleLogger) Warn(message string) {
	if l.level <= LogLevelWarn {
		fmt.Println("WARN: " + message)
	}
}

func (l *ConsoleLogger) Warnf(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		fmt.Println(fmt.Sprintf("WARN: "+format, args...))
	}
}

func (l *ConsoleLogger) Error(message string) {
	if l.level <= LogLevelError {
		fmt.Println("ERROR: " + message)
	}
}

func (l *ConsoleLogger) Errorf(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		fmt.Println(fmt.Sprintf("ERROR: "+format, args...))
	}
}

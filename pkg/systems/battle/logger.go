package battle

// Logger is the minimal logging surface the battle system needs. It is defined
// here (rather than imported from the engine) so the package stays reusable
// across games and easy to test. The engine's logger.Logger satisfies it, but
// any implementation - including none - works.
type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

// nopLogger is a Logger that discards everything. It is the default so the
// manager never has to nil-check its logger.
type nopLogger struct{}

func (nopLogger) Debugf(string, ...interface{}) {}
func (nopLogger) Warnf(string, ...interface{})  {}

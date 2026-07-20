package battle

// Config holds the tunable parameters for the battle system. It exists so the
// package has no dependency on any engine-global configuration: the host game
// passes in exactly the values it wants. Zero-valued fields are replaced with
// sane defaults by normalize(), so callers may set only the fields they care
// about.
type Config struct {
	// ActionQueueSize is the buffer size of the action queue channel.
	ActionQueueSize int

	// TimerChargeRate multiplies delta time when charging entity action timers
	// (1.0 = one unit of charge per second).
	TimerChargeRate float64

	// DamageEffectDuration is how long floating damage/heal numbers live, in
	// seconds.
	DamageEffectDuration float64

	// Logger receives diagnostic messages. If nil, logging is silently
	// discarded (see nopLogger).
	Logger Logger
}

// Default values applied to any unset (non-positive / nil) Config field.
const (
	defaultActionQueueSize      = 100
	defaultTimerChargeRate      = 0.33
	defaultDamageEffectDuration = 2.0
)

// DefaultConfig returns a Config populated with reasonable defaults. It is a
// convenient starting point that callers can copy and tweak.
func DefaultConfig() Config {
	return Config{
		ActionQueueSize:      defaultActionQueueSize,
		TimerChargeRate:      defaultTimerChargeRate,
		DamageEffectDuration: defaultDamageEffectDuration,
		Logger:               nopLogger{},
	}
}

// normalize returns a copy of c with any unset field replaced by its default,
// guaranteeing the manager always has usable values and a non-nil Logger.
func (c Config) normalize() Config {
	if c.ActionQueueSize <= 0 {
		c.ActionQueueSize = defaultActionQueueSize
	}
	if c.TimerChargeRate <= 0 {
		c.TimerChargeRate = defaultTimerChargeRate
	}
	if c.DamageEffectDuration <= 0 {
		c.DamageEffectDuration = defaultDamageEffectDuration
	}
	if c.Logger == nil {
		c.Logger = nopLogger{}
	}
	return c
}

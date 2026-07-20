package battle

import "github.com/cstevenson98/gowasm-engine/pkg/types"

// BattleEntity represents an entity that can participate in battle.
//
// This interface is defined here, in the package that consumes it, following
// the Go idiom that the consumer owns the interface. Game entities (a Player,
// an Enemy) implement it to join a battle; they depend on this package rather
// than the engine's shared types package depending on battle concepts.
type BattleEntity interface {
	// GetActionTimer returns the entity's action timer.
	GetActionTimer() *ActionTimer

	// ChargeTimer charges the action timer by deltaTime.
	ChargeTimer(deltaTime float64)

	// ResetTimer resets the action timer to 0.
	ResetTimer()

	// IsReady returns true if the entity can take an action (timer >= 1.0).
	IsReady() bool

	// GetStats returns the entity's battle stats.
	GetStats() *EntityStats

	// SelectAction returns the action this entity wants to perform, or nil if
	// no action should be taken.
	SelectAction() *Action

	// GetID returns the entity's unique identifier.
	GetID() string

	// GetMover returns the mover component for position access.
	GetMover() types.Mover
}

// EntityStats represents the battle statistics of an entity.
type EntityStats struct {
	HP    int
	MaxHP int
	Speed float64 // Charge rate multiplier (1.0 = normal speed)
}

// ActionTimer represents an entity's action timer.
type ActionTimer struct {
	Current    float64 // 0.0 to 1.0
	ChargeRate float64 // How fast it charges (1.0 = 1.0 per second)
	IsCharging bool    // Whether timer is currently charging
}

// NewActionTimer creates a new action timer with default values.
func NewActionTimer() *ActionTimer {
	return &ActionTimer{
		Current:    0.0,
		ChargeRate: 1.0, // 1.0 per second
		IsCharging: true,
	}
}

// Charge adds deltaTime to the timer if it's charging.
func (at *ActionTimer) Charge(deltaTime float64) {
	if at.IsCharging {
		at.Current += deltaTime * at.ChargeRate
		if at.Current > 1.0 {
			at.Current = 1.0
		}
	}
}

// Reset sets the timer back to 0.
func (at *ActionTimer) Reset() {
	at.Current = 0.0
}

// IsFull returns true if the timer has reached 1.0.
func (at *ActionTimer) IsFull() bool {
	return at.Current >= 1.0
}

// SetCharging sets whether the timer is currently charging.
func (at *ActionTimer) SetCharging(charging bool) {
	at.IsCharging = charging
}

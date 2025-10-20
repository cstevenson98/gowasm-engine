package types

import (
	"math/rand"
	"time"
)

// BattleEntity represents an entity that can participate in battle
type BattleEntity interface {
	// GetActionTimer returns the entity's action timer
	GetActionTimer() *ActionTimer

	// ChargeTimer charges the action timer by deltaTime
	ChargeTimer(deltaTime float64)

	// ResetTimer resets the action timer to 0
	ResetTimer()

	// IsReady returns true if the entity can take an action (timer >= 1.0)
	IsReady() bool

	// GetStats returns the entity's battle stats
	GetStats() *EntityStats

	// SelectAction returns the action this entity wants to perform
	// Returns nil if no action should be taken
	SelectAction() *Action

	// GetID returns the entity's unique identifier
	GetID() string

	// GetMover returns the mover component for position access
	GetMover() Mover
}

// EntityStats represents the battle statistics of an entity
type EntityStats struct {
	HP    int
	MaxHP int
	Speed float64 // Charge rate multiplier (1.0 = normal speed)
}

// ActionTimer represents an entity's action timer
type ActionTimer struct {
	Current    float64 // 0.0 to 1.0
	ChargeRate float64 // How fast it charges (1.0 = 1.0 per second)
	IsCharging bool    // Whether timer is currently charging
}

// NewActionTimer creates a new action timer with default values
func NewActionTimer() *ActionTimer {
	return &ActionTimer{
		Current:    0.0,
		ChargeRate: 1.0, // 1.0 per second
		IsCharging: true,
	}
}

// Charge adds deltaTime to the timer if it's charging
func (at *ActionTimer) Charge(deltaTime float64) {
	if at.IsCharging {
		at.Current += deltaTime * at.ChargeRate
		if at.Current > 1.0 {
			at.Current = 1.0
		}
	}
}

// Reset sets the timer back to 0
func (at *ActionTimer) Reset() {
	at.Current = 0.0
}

// IsFull returns true if the timer has reached 1.0
func (at *ActionTimer) IsFull() bool {
	return at.Current >= 1.0
}

// SetCharging sets whether the timer is currently charging
func (at *ActionTimer) SetCharging(charging bool) {
	at.IsCharging = charging
}

// ActionType represents the type of action that can be performed
type ActionType int

const (
	ActionAttack ActionType = iota
	ActionDefend
	ActionItem
	ActionRun
	ActionHaunt // Enemy-specific action
)

// String returns the string representation of the action type
func (at ActionType) String() string {
	switch at {
	case ActionAttack:
		return "Attack"
	case ActionDefend:
		return "Defend"
	case ActionItem:
		return "Item"
	case ActionRun:
		return "Run"
	case ActionHaunt:
		return "Haunt"
	default:
		return "Unknown"
	}
}

// Action represents a battle action to be performed
type Action struct {
	Type              ActionType
	Actor             BattleEntity
	Target            BattleEntity
	Damage            int
	AnimationDuration float64
	Description       string
}

// NewAction creates a new action
func NewAction(actionType ActionType, actor, target BattleEntity, damage int, duration float64, description string) *Action {
	return &Action{
		Type:              actionType,
		Actor:             actor,
		Target:            target,
		Damage:            damage,
		AnimationDuration: duration,
		Description:       description,
	}
}

// GetRandomDamage returns a random damage value between min and max (inclusive)
func GetRandomDamage(min, max int) int {
	if min >= max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

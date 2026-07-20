package battle

import (
	"math/rand"
	"time"
)

// ActionType represents the type of action that can be performed.
type ActionType int

const (
	ActionAttack ActionType = iota
	ActionDefend
	ActionItem
	ActionRun
	ActionHaunt // Enemy-specific action
)

// String returns the string representation of the action type.
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

// Action represents a battle action to be performed.
type Action struct {
	Type              ActionType
	Actor             BattleEntity
	Target            BattleEntity
	Damage            int
	AnimationDuration float64
	Description       string
}

// NewAction creates a new action.
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

// GetRandomDamage returns a random damage value between min and max (inclusive).
func GetRandomDamage(min, max int) int {
	if min >= max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

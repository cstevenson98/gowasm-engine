package battle

import (
	"sync"
)

// ActionQueue is a bounded FIFO of battle actions. It is drained synchronously
// by BattleManager.Update on the main loop (the battle system no longer runs a
// background goroutine, so it is safe to touch ECS/engine state from actions).
type ActionQueue struct {
	mu       sync.Mutex
	actions  []*Action
	capacity int
}

// NewActionQueue creates a new action queue bounded to bufferSize entries.
func NewActionQueue(bufferSize int) *ActionQueue {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &ActionQueue{
		actions:  make([]*Action, 0, bufferSize),
		capacity: bufferSize,
	}
}

// Enqueue appends an action, returning false if the queue is full.
func (aq *ActionQueue) Enqueue(action *Action) bool {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	if len(aq.actions) >= aq.capacity {
		return false
	}
	aq.actions = append(aq.actions, action)
	return true
}

// Dequeue removes and returns the oldest action; ok is false when empty.
func (aq *ActionQueue) Dequeue() (*Action, bool) {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	if len(aq.actions) == 0 {
		return nil, false
	}
	action := aq.actions[0]
	aq.actions = aq.actions[1:]
	return action, true
}

// Size returns the current number of queued actions.
func (aq *ActionQueue) Size() int {
	aq.mu.Lock()
	defer aq.mu.Unlock()
	return len(aq.actions)
}

// AvailableActions returns the list of actions available to a player
func AvailableActions() []ActionType {
	return []ActionType{
		ActionAttack,
		ActionDefend,
		ActionItem,
		ActionRun,
	}
}

// AvailableEnemyActions returns the list of actions available to enemies
func AvailableEnemyActions() []ActionType {
	return []ActionType{
		ActionHaunt,
	}
}

// CreatePlayerAction creates an action for a player based on the selected action type
func CreatePlayerAction(actionType ActionType, actor, target BattleEntity) *Action {
	switch actionType {
	case ActionAttack:
		// Simple attack: 5-8 damage
		damage := GetRandomDamage(5, 8)
		return NewAction(
			actionType,
			actor,
			target,
			damage,
			1.0, // 1 second animation
			"attacks",
		)
	case ActionDefend:
		// Defend: no damage, but reduces incoming damage
		return NewAction(
			actionType,
			actor,
			target,
			0,
			0.5, // 0.5 second animation
			"defends",
		)
	case ActionItem:
		// Item: heal for 10-15 HP
		heal := GetRandomDamage(10, 15)
		return NewAction(
			actionType,
			actor,
			actor, // Target self for healing
			-heal, // Negative damage = healing
			1.0,
			"uses an item",
		)
	case ActionRun:
		// Run: attempt to flee (no damage)
		return NewAction(
			actionType,
			actor,
			nil, // No target for running
			0,
			0.5,
			"attempts to run",
		)
	default:
		return nil
	}
}

// CreateEnemyAction creates an action for an enemy (random selection)
func CreateEnemyAction(actor, target BattleEntity) *Action {
	// For now, enemies only have the "Haunt" action
	// In the future, this could be expanded with AI logic
	damage := GetRandomDamage(9, 12) // Haunt attack: 9-12 damage
	return NewAction(
		ActionHaunt,
		actor,
		target,
		damage,
		1.2, // 1.2 second animation (slightly longer than player attack)
		"haunts",
	)
}

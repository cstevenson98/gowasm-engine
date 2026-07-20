package battle

import (
	"sync"
)

// ActionQueue manages the queue of battle actions using channels
type ActionQueue struct {
	actions    chan *Action
	processing sync.Mutex
	closed     bool
}

// NewActionQueue creates a new action queue with a buffered channel
func NewActionQueue(bufferSize int) *ActionQueue {
	return &ActionQueue{
		actions: make(chan *Action, bufferSize),
		closed:  false,
	}
}

// Enqueue adds an action to the queue
func (aq *ActionQueue) Enqueue(action *Action) bool {
	if aq.closed {
		return false
	}

	select {
	case aq.actions <- action:
		return true
	default:
		// Channel is full, action is dropped
		return false
	}
}

// Dequeue removes and returns an action from the queue
func (aq *ActionQueue) Dequeue() (*Action, bool) {
	action, ok := <-aq.actions
	return action, ok
}

// Close closes the action queue
func (aq *ActionQueue) Close() {
	aq.processing.Lock()
	defer aq.processing.Unlock()

	if !aq.closed {
		close(aq.actions)
		aq.closed = true
	}
}

// IsClosed returns true if the queue is closed
func (aq *ActionQueue) IsClosed() bool {
	return aq.closed
}

// Size returns the current number of actions in the queue
func (aq *ActionQueue) Size() int {
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

//go:build js

package gameobject

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/mover"
	"github.com/cstevenson98/gowasm-engine/pkg/sprite"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Enemy represents an enemy character in battle.
// It embeds BaseGameObject to inherit common GameObject functionality.
type Enemy struct {
	*BaseGameObject

	// Enemy-specific fields
	size types.Vector2

	// Battle system
	actionTimer *types.ActionTimer
	stats       *types.EntityStats
	mu          sync.Mutex
}

// NewEnemy creates a new enemy game object
func NewEnemy(position, size types.Vector2, texturePath string) *Enemy {
	// Create sprite with ghost animation (2 rows, 3 columns)
	enemySprite := sprite.NewSpriteSheet(
		texturePath,
		sprite.Vector2{X: size.X, Y: size.Y},
		3, // 3 columns
		2, // 2 rows
	)

	// Create static mover (enemies don't move in battle)
	enemyMover := mover.NewBasicMover(position, types.Vector2{X: 0, Y: 0}, config.Global.Screen.Width, config.Global.Screen.Height)

	// Create state
	enemyState := types.ObjectState{
		ID:       "enemy",
		Position: position,
		Visible:  true,
		Frame:    0,
	}

	// Initialize BaseGameObject
	baseGameObject := NewBaseGameObject(enemySprite, enemyMover, enemyState)

	logger.Logger.Debugf("Created Enemy at position (%.2f, %.2f)", position.X, position.Y)

	return &Enemy{
		BaseGameObject: baseGameObject,
		size:           size,
		actionTimer:    types.NewActionTimer(),
		stats: &types.EntityStats{
			HP:    80, // Will be overridden by config
			MaxHP: 80,
			Speed: 1.0,
		},
	}
}

// Update updates the enemy's state (overrides BaseGameObject.Update)
func (e *Enemy) Update(deltaTime float64) {
	// Enemy doesn't do much in battle for now
	// Just update sprite animation
	if e.GetSprite() != nil {
		e.GetSprite().Update(deltaTime)
	}
}

// BattleEntity interface implementation

// GetActionTimer returns the enemy's action timer
func (e *Enemy) GetActionTimer() *types.ActionTimer {
	return e.actionTimer
}

// ChargeTimer charges the action timer by deltaTime
func (e *Enemy) ChargeTimer(deltaTime float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.actionTimer.Charge(deltaTime)
}

// ResetTimer resets the action timer to 0
func (e *Enemy) ResetTimer() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.actionTimer.Reset()
}

// IsReady returns true if the enemy can take an action
func (e *Enemy) IsReady() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.actionTimer.IsFull()
}

// GetStats returns the enemy's battle stats
func (e *Enemy) GetStats() *types.EntityStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stats
}

// SelectAction returns the enemy's selected action (random for now)
func (e *Enemy) SelectAction() *types.Action {
	e.mu.Lock()
	defer e.mu.Unlock()

	// For now, enemies always use the "Haunt" action
	// In the future, this could be expanded with AI logic
	// We need a target - this will be set by the battle manager
	return nil // Battle manager will create the action with proper target
}

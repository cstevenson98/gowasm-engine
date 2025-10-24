//go:build js

package gameobject

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/mover"
	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Enemy represents an enemy character in battle
type Enemy struct {
	id       string
	position types.Vector2
	size     types.Vector2
	sprite   types.Sprite
	mover    types.Mover
	visible  bool

	// Battle system
	actionTimer *types.ActionTimer
	stats       *types.EntityStats
	mu          sync.Mutex
}

// NewEnemy creates a new enemy game object
func NewEnemy(position, size types.Vector2, texturePath string) *Enemy {
	enemy := &Enemy{
		id:          "enemy",
		position:    position,
		size:        size,
		visible:     true,
		actionTimer: types.NewActionTimer(),
		stats: &types.EntityStats{
			HP:    80, // Will be overridden by config
			MaxHP: 80,
			Speed: 1.0,
		},
	}

	// Create sprite with ghost animation (2 rows, 3 columns)
	enemySprite := sprite.NewSpriteSheet(
		texturePath,
		sprite.Vector2{X: size.X, Y: size.Y},
		3, // 3 columns
		2, // 2 rows
	)
	enemy.sprite = enemySprite

	// Create static mover (enemies don't move in battle)
	enemyMover := mover.NewBasicMover(position, types.Vector2{X: 0, Y: 0}, config.Global.Screen.Width, config.Global.Screen.Height)
	enemy.mover = enemyMover

	logger.Logger.Debugf("Created Enemy at position (%.2f, %.2f)", position.X, position.Y)

	return enemy
}

// Update updates the enemy's state
func (e *Enemy) Update(deltaTime float64) {
	// Enemy doesn't do much in battle for now
	// Just update sprite animation
	if e.sprite != nil {
		e.sprite.Update(deltaTime)
	}
}

// GetSprite returns the enemy's sprite
func (e *Enemy) GetSprite() types.Sprite {
	return e.sprite
}

// GetMover returns the enemy's mover
func (e *Enemy) GetMover() types.Mover {
	return e.mover
}

// SetVisibility sets the enemy's visibility
func (e *Enemy) SetVisibility(visible bool) {
	e.visible = visible
}

// IsVisible returns whether the enemy is visible
func (e *Enemy) IsVisible() bool {
	return e.visible
}

// GetID returns the enemy's unique identifier
func (e *Enemy) GetID() string {
	return e.id
}

// GetState returns the enemy's current state
func (e *Enemy) GetState() *types.ObjectState {
	return &types.ObjectState{
		ID:       e.id,
		Position: e.position,
		Visible:  e.visible,
		Frame:    0, // Enemies don't animate for now
	}
}

// SetState sets the enemy's state
func (e *Enemy) SetState(state types.ObjectState) {
	e.position = state.Position
	e.visible = state.Visible
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

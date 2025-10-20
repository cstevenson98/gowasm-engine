package battle

import (
	"context"
	"sync"
	"time"

	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/types"
)

// BattleManager manages the battle system including action queue processing
type BattleManager struct {
	actionQueue    *ActionQueue
	entities       []types.BattleEntity
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	processingDone chan bool
	effectManager  *EffectManager
}

// NewBattleManager creates a new battle manager
func NewBattleManager() *BattleManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &BattleManager{
		actionQueue:    NewActionQueue(config.Global.Battle.ActionQueueSize),
		entities:       make([]types.BattleEntity, 0),
		mu:             sync.RWMutex{},
		ctx:            ctx,
		cancel:         cancel,
		processingDone: make(chan bool, 1),
		effectManager:  NewEffectManager(),
	}
}

// AddEntity adds a battle entity to the manager
func (bm *BattleManager) AddEntity(entity types.BattleEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.entities = append(bm.entities, entity)
	logger.Logger.Debugf("Added entity %s to battle manager", entity.GetID())
}

// RemoveEntity removes a battle entity from the manager
func (bm *BattleManager) RemoveEntity(entity types.BattleEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i, e := range bm.entities {
		if e.GetID() == entity.GetID() {
			bm.entities = append(bm.entities[:i], bm.entities[i+1:]...)
			logger.Logger.Debugf("Removed entity %s from battle manager", entity.GetID())
			return
		}
	}
}

// StartProcessing starts the action queue processing goroutine
func (bm *BattleManager) StartProcessing() {
	go bm.processActions()
	logger.Logger.Debugf("Started battle manager action processing")
}

// StopProcessing stops the action queue processing
func (bm *BattleManager) StopProcessing() {
	bm.cancel()
	bm.actionQueue.Close()

	// Wait for processing to complete
	select {
	case <-bm.processingDone:
		logger.Logger.Debugf("Battle manager processing stopped")
	case <-time.After(5 * time.Second):
		logger.Logger.Warnf("Battle manager stop timeout")
	}
}

// EnqueueAction adds an action to the queue
func (bm *BattleManager) EnqueueAction(action *types.Action) bool {
	if action == nil {
		return false
	}

	success := bm.actionQueue.Enqueue(action)
	if success {
		logger.Logger.Debugf("Enqueued action: %s", action.Description)
	} else {
		logger.Logger.Warnf("Failed to enqueue action: %s", action.Description)
	}
	return success
}

// Update updates the battle manager (charges timers, processes actions)
func (bm *BattleManager) Update(deltaTime float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Always charge timers for all entities (no animation blocking)
	bm.chargeAllTimers(deltaTime)

	// Check for entities ready to act
	bm.checkForReadyEntities()
}

// IsAnimating returns false (no animation blocking in dynamic battle)
func (bm *BattleManager) IsAnimating() bool {
	return false
}

// GetEntities returns a copy of the entities list
func (bm *BattleManager) GetEntities() []types.BattleEntity {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	entities := make([]types.BattleEntity, len(bm.entities))
	copy(entities, bm.entities)
	return entities
}

// GetEffectManager returns the effect manager
func (bm *BattleManager) GetEffectManager() *EffectManager {
	return bm.effectManager
}

// processActions runs in a goroutine to process the action queue
func (bm *BattleManager) processActions() {
	defer func() {
		bm.processingDone <- true
	}()

	for {
		select {
		case <-bm.ctx.Done():
			logger.Logger.Debugf("Battle manager processing stopped by context")
			return
		case action, ok := <-bm.actionQueue.actions:
			if !ok {
				logger.Logger.Debugf("Action queue closed")
				return
			}

			bm.processAction(action)
		}
	}
}

// processAction executes a single action
func (bm *BattleManager) processAction(action *types.Action) {
	logger.Logger.Debugf("Processing action: %s %s", action.Actor.GetID(), action.Description)

	// Execute the action (no animation blocking)
	bm.executeAction(action)

	// Reset the actor's timer
	action.Actor.ResetTimer()
}

// executeAction performs the actual action effects
func (bm *BattleManager) executeAction(action *types.Action) {
	switch action.Type {
	case types.ActionAttack, types.ActionHaunt:
		bm.executeDamage(action)
	case types.ActionDefend:
		bm.executeDefend(action)
	case types.ActionItem:
		bm.executeHeal(action)
	case types.ActionRun:
		bm.executeRun(action)
	default:
		logger.Logger.Warnf("Unknown action type: %v", action.Type)
	}
}

// executeDamage applies damage to the target
func (bm *BattleManager) executeDamage(action *types.Action) {
	if action.Target == nil {
		logger.Logger.Warnf("Damage action has no target")
		return
	}

	stats := action.Target.GetStats()
	stats.HP -= action.Damage
	if stats.HP < 0 {
		stats.HP = 0
	}

	// Create damage effect
	// Get target position (assuming it has a mover)
	if mover := action.Target.GetMover(); mover != nil {
		pos := mover.GetPosition()
		// Offset slightly above the entity
		effectPos := types.Vector2{X: pos.X, Y: pos.Y - 20}
		damageEffect := NewDamageEffect(effectPos, action.Damage, 2.0, false) // 2 second duration
		bm.effectManager.AddEffect(damageEffect)
	}

	logger.Logger.Debugf("%s deals %d damage to %s (HP: %d/%d)",
		action.Actor.GetID(), action.Damage, action.Target.GetID(),
		stats.HP, stats.MaxHP)
}

// executeDefend applies defense effect
func (bm *BattleManager) executeDefend(action *types.Action) {
	logger.Logger.Debugf("%s defends", action.Actor.GetID())
	// Defense logic would go here (e.g., set defense flag)
}

// executeHeal applies healing to the target
func (bm *BattleManager) executeHeal(action *types.Action) {
	if action.Target == nil {
		logger.Logger.Warnf("Heal action has no target")
		return
	}

	stats := action.Target.GetStats()
	healAmount := -action.Damage // Negative damage = healing
	stats.HP += healAmount
	if stats.HP > stats.MaxHP {
		stats.HP = stats.MaxHP
	}

	// Create healing effect
	if mover := action.Target.GetMover(); mover != nil {
		pos := mover.GetPosition()
		// Offset slightly above the entity
		effectPos := types.Vector2{X: pos.X, Y: pos.Y - 20}
		healEffect := NewDamageEffect(effectPos, healAmount, 2.0, true) // 2 second duration, healing
		bm.effectManager.AddEffect(healEffect)
	}

	logger.Logger.Debugf("%s heals %d HP to %s (HP: %d/%d)",
		action.Actor.GetID(), healAmount, action.Target.GetID(),
		stats.HP, stats.MaxHP)
}

// executeRun handles run action
func (bm *BattleManager) executeRun(action *types.Action) {
	logger.Logger.Debugf("%s attempts to run", action.Actor.GetID())
	// Run logic would go here (e.g., chance to escape)
}

// chargeAllTimers charges all entity timers
func (bm *BattleManager) chargeAllTimers(deltaTime float64) {
	chargeRate := config.Global.Battle.TimerChargeRate
	for _, entity := range bm.entities {
		// Apply charge rate to deltaTime
		entity.ChargeTimer(deltaTime * chargeRate)
	}
}

// checkForReadyEntities checks if any entities are ready to act
func (bm *BattleManager) checkForReadyEntities() {
	for _, entity := range bm.entities {
		if entity.IsReady() {
			// Only automatically handle non-player entities (enemies)
			// Player actions are handled by menu selection
			if entity.GetID() != "Player" {
				action := entity.SelectAction()
				if action != nil {
					bm.EnqueueAction(action)
				} else {
					// Handle entities that don't select their own actions (like enemies)
					// Find a target for the entity (for now, just find any other entity)
					var target types.BattleEntity
					for _, otherEntity := range bm.entities {
						if otherEntity.GetID() != entity.GetID() {
							target = otherEntity
							break
						}
					}

					if target != nil {
						// Create enemy action (for now, always Haunt attack)
						action := CreateEnemyAction(entity, target)
						if action != nil {
							bm.EnqueueAction(action)
						}
					}
				}
			}
		}
	}
}

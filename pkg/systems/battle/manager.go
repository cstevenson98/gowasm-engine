package battle

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BattleManager manages the battle system including action queue processing.
// Actions are processed synchronously on the main loop (see Update); there is
// no background goroutine, so action handlers may safely touch ECS/engine state.
type BattleManager struct {
	actionQueue   *ActionQueue
	entities      []BattleEntity
	mu            sync.RWMutex
	effectManager *EffectManager
	cfg           Config
	log           Logger
}

// NewBattleManager creates a new battle manager configured by cfg. All engine
// coupling (timings, queue size, logging) enters through Config, so the same
// system can be dropped into any game. Unset Config fields fall back to
// sensible defaults - pass battle.Config{} for an all-default manager.
func NewBattleManager(cfg Config) *BattleManager {
	cfg = cfg.normalize()
	return &BattleManager{
		actionQueue:   NewActionQueue(cfg.ActionQueueSize),
		entities:      make([]BattleEntity, 0),
		effectManager: NewEffectManager(),
		cfg:           cfg,
		log:           cfg.Logger,
	}
}

// AddEntity adds a battle entity to the manager
func (bm *BattleManager) AddEntity(entity BattleEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.entities = append(bm.entities, entity)
	bm.log.Debugf("Added entity %s to battle manager", entity.GetID())
}

// RemoveEntity removes a battle entity from the manager
func (bm *BattleManager) RemoveEntity(entity BattleEntity) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for i, e := range bm.entities {
		if e.GetID() == entity.GetID() {
			bm.entities = append(bm.entities[:i], bm.entities[i+1:]...)
			bm.log.Debugf("Removed entity %s from battle manager", entity.GetID())
			return
		}
	}
}

// EnqueueAction adds an action to the queue
func (bm *BattleManager) EnqueueAction(action *Action) bool {
	if action == nil {
		return false
	}

	success := bm.actionQueue.Enqueue(action)
	if success {
		bm.log.Debugf("Enqueued action: %s", action.Description)
	} else {
		bm.log.Warnf("Failed to enqueue action: %s", action.Description)
	}
	return success
}

// Update charges timers, lets ready entities act, and processes all queued
// actions synchronously. Called once per frame from the game loop.
func (bm *BattleManager) Update(deltaTime float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Always charge timers for all entities (no animation blocking)
	bm.chargeAllTimers(deltaTime)

	// Check for entities ready to act (may enqueue enemy actions)
	bm.checkForReadyEntities()

	// Drain and process the action queue on the main loop.
	for {
		action, ok := bm.actionQueue.Dequeue()
		if !ok {
			break
		}
		bm.processAction(action)
	}
}

// IsAnimating returns false (no animation blocking in dynamic battle)
func (bm *BattleManager) IsAnimating() bool {
	return false
}

// GetEntities returns a copy of the entities list
func (bm *BattleManager) GetEntities() []BattleEntity {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	entities := make([]BattleEntity, len(bm.entities))
	copy(entities, bm.entities)
	return entities
}

// GetEffectManager returns the effect manager
func (bm *BattleManager) GetEffectManager() *EffectManager {
	return bm.effectManager
}

// processAction executes a single action
func (bm *BattleManager) processAction(action *Action) {
	bm.log.Debugf("Processing action: %s %s", action.Actor.GetID(), action.Description)

	// Execute the action (no animation blocking)
	bm.executeAction(action)

	// Reset the actor's timer
	action.Actor.ResetTimer()
}

// executeAction performs the actual action effects
func (bm *BattleManager) executeAction(action *Action) {
	switch action.Type {
	case ActionAttack, ActionHaunt:
		bm.executeDamage(action)
	case ActionDefend:
		bm.executeDefend(action)
	case ActionItem:
		bm.executeHeal(action)
	case ActionRun:
		bm.executeRun(action)
	default:
		bm.log.Warnf("Unknown action type: %v", action.Type)
	}
}

// executeDamage applies damage to the target
func (bm *BattleManager) executeDamage(action *Action) {
	if action.Target == nil {
		bm.log.Warnf("Damage action has no target")
		return
	}

	stats := action.Target.GetStats()
	stats.HP -= action.Damage
	if stats.HP < 0 {
		stats.HP = 0
	}

	// Create damage effect, offset slightly above the target.
	pos := action.Target.GetPosition()
	effectPos := types.Vector2{X: pos.X, Y: pos.Y - 20}
	damageEffect := NewDamageEffect(effectPos, action.Damage, bm.cfg.DamageEffectDuration, false)
	bm.effectManager.AddEffect(damageEffect)

	bm.log.Debugf("%s deals %d damage to %s (HP: %d/%d)",
		action.Actor.GetID(), action.Damage, action.Target.GetID(),
		stats.HP, stats.MaxHP)
}

// executeDefend applies defense effect
func (bm *BattleManager) executeDefend(action *Action) {
	bm.log.Debugf("%s defends", action.Actor.GetID())
	// Defense logic would go here (e.g., set defense flag)
}

// executeHeal applies healing to the target
func (bm *BattleManager) executeHeal(action *Action) {
	if action.Target == nil {
		bm.log.Warnf("Heal action has no target")
		return
	}

	stats := action.Target.GetStats()
	healAmount := -action.Damage // Negative damage = healing
	stats.HP += healAmount
	if stats.HP > stats.MaxHP {
		stats.HP = stats.MaxHP
	}

	// Create healing effect, offset slightly above the target.
	pos := action.Target.GetPosition()
	effectPos := types.Vector2{X: pos.X, Y: pos.Y - 20}
	healEffect := NewDamageEffect(effectPos, healAmount, bm.cfg.DamageEffectDuration, true)
	bm.effectManager.AddEffect(healEffect)

	bm.log.Debugf("%s heals %d HP to %s (HP: %d/%d)",
		action.Actor.GetID(), healAmount, action.Target.GetID(),
		stats.HP, stats.MaxHP)
}

// executeRun handles run action
func (bm *BattleManager) executeRun(action *Action) {
	bm.log.Debugf("%s attempts to run", action.Actor.GetID())
	// Run logic would go here (e.g., chance to escape)
}

// chargeAllTimers charges all entity timers
func (bm *BattleManager) chargeAllTimers(deltaTime float64) {
	chargeRate := bm.cfg.TimerChargeRate
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
					var target BattleEntity
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

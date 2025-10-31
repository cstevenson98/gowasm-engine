package battle

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// DamageEffect represents a visual damage/healing effect
type DamageEffect struct {
	Position  types.Vector2
	Value     int
	Duration  float64
	Elapsed   float64
	IsHealing bool
	mu        sync.RWMutex
}

// NewDamageEffect creates a new damage effect
func NewDamageEffect(position types.Vector2, value int, duration float64, isHealing bool) *DamageEffect {
	return &DamageEffect{
		Position:  position,
		Value:     value,
		Duration:  duration,
		Elapsed:   0.0,
		IsHealing: isHealing,
	}
}

// Update updates the effect's elapsed time
func (de *DamageEffect) Update(deltaTime float64) {
	de.mu.Lock()
	defer de.mu.Unlock()

	de.Elapsed += deltaTime
}

// IsFinished returns true if the effect has completed
func (de *DamageEffect) IsFinished() bool {
	de.mu.RLock()
	defer de.mu.RUnlock()

	return de.Elapsed >= de.Duration
}

// GetAlpha returns the alpha value for fading (1.0 to 0.0)
func (de *DamageEffect) GetAlpha() float32 {
	de.mu.RLock()
	defer de.mu.RUnlock()

	if de.Elapsed >= de.Duration {
		return 0.0
	}

	// Linear fade from 1.0 to 0.0
	progress := de.Elapsed / de.Duration
	return float32(1.0 - progress)
}

// GetPosition returns the current position (may include floating animation)
func (de *DamageEffect) GetPosition() types.Vector2 {
	de.mu.RLock()
	defer de.mu.RUnlock()

	// Simple floating animation - move up over time
	floatOffset := de.Elapsed * 30.0 // 30 pixels per second upward
	return types.Vector2{
		X: de.Position.X,
		Y: de.Position.Y - floatOffset,
	}
}

// GetValue returns the damage/healing value
func (de *DamageEffect) GetValue() int {
	de.mu.RLock()
	defer de.mu.RUnlock()

	return de.Value
}

// IsHealingEffect returns true if this is a healing effect
func (de *DamageEffect) IsHealingEffect() bool {
	de.mu.RLock()
	defer de.mu.RUnlock()

	return de.IsHealing
}

// EffectManager manages all active effects
type EffectManager struct {
	effects []*DamageEffect
	mu      sync.RWMutex
}

// NewEffectManager creates a new effect manager
func NewEffectManager() *EffectManager {
	return &EffectManager{
		effects: make([]*DamageEffect, 0),
	}
}

// AddEffect adds a new damage effect
func (em *EffectManager) AddEffect(effect *DamageEffect) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.effects = append(em.effects, effect)
}

// Update updates all effects and removes finished ones
func (em *EffectManager) Update(deltaTime float64) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Update all effects
	for _, effect := range em.effects {
		effect.Update(deltaTime)
	}

	// Remove finished effects
	activeEffects := make([]*DamageEffect, 0)
	for _, effect := range em.effects {
		if !effect.IsFinished() {
			activeEffects = append(activeEffects, effect)
		}
	}
	em.effects = activeEffects
}

// GetActiveEffects returns a copy of all active effects
func (em *EffectManager) GetActiveEffects() []*DamageEffect {
	em.mu.RLock()
	defer em.mu.RUnlock()

	effects := make([]*DamageEffect, len(em.effects))
	copy(effects, em.effects)
	return effects
}

// ClearAllEffects removes all effects
func (em *EffectManager) ClearAllEffects() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.effects = make([]*DamageEffect, 0)
}

// GetEffectCount returns the number of active effects
func (em *EffectManager) GetEffectCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return len(em.effects)
}

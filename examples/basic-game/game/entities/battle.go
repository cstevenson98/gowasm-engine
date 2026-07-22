package entities

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/mover"
	"github.com/cstevenson98/gowasm-engine/pkg/systems/battle"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// Participant adapts a battle combatant to the battle system's BattleEntity
// interface. It holds the combatant's timer/stats and a static position (battle
// combatants don't move), decoupled from the ECS render entity. This keeps the
// battle manager working unchanged; a later pass folds stats/timers into
// components and removes the manager's goroutine.
type Participant struct {
	id       string
	timer    *battle.ActionTimer
	stats    *battle.EntityStats
	mover    types.Mover
	isPlayer bool
	mu       sync.Mutex
}

// NewParticipant creates a battle participant. id must be "Player" for the
// player (the battle manager keys player-vs-enemy behaviour on this).
func NewParticipant(id string, hp, maxHP int, position types.Vector2, isPlayer bool) *Participant {
	return &Participant{
		id:       id,
		timer:    battle.NewActionTimer(),
		stats:    &battle.EntityStats{HP: hp, MaxHP: maxHP, Speed: 1.0},
		mover:    mover.NewBasicMover(position, types.Vector2{}, 0, 0),
		isPlayer: isPlayer,
	}
}

func (p *Participant) GetID() string { return p.id }

func (p *Participant) GetActionTimer() *battle.ActionTimer { return p.timer }

func (p *Participant) ChargeTimer(deltaTime float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timer.Charge(deltaTime)
}

func (p *Participant) ResetTimer() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timer.Reset()
}

func (p *Participant) IsReady() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.timer.IsFull()
}

func (p *Participant) GetStats() *battle.EntityStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

// SelectAction returns nil: the player acts via the battle menu, and the battle
// manager synthesises enemy actions.
func (p *Participant) SelectAction() *battle.Action { return nil }

func (p *Participant) GetMover() types.Mover { return p.mover }

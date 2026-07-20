// Package battle implements a real-time, timer-based combat system.
//
// Combat here is not strictly turn-based. Every participant charges an action
// timer continuously; when a timer fills, that participant may act. This gives
// an "active time battle" feel where faster entities act more often.
//
// # Participants
//
// Any object implementing types.BattleEntity can join a battle - in this
// engine, gameobject.Player and gameobject.Enemy do. An entity exposes a
// charging types.ActionTimer, battle stats (HP, speed), and a way to select the
// action it wants to perform.
//
// # The manager
//
// BattleManager is the orchestrator. Each frame it charges every entity's
// timer and, for any entity that is ready, produces a types.Action. Player
// actions come from menu selection; enemy actions are chosen automatically.
// Actions are pushed onto an ActionQueue (a buffered channel) and applied by a
// background goroutine, which resolves attacks, defends, items, and run
// attempts, adjusting the target's HP.
//
// # Effects
//
// Resolving an action can spawn a transient visual through the EffectManager -
// for example a floating, fading DamageEffect showing damage dealt or HP
// healed above the target. Effects age out on their own and are purged once
// finished.
//
// Action construction helpers (CreatePlayerAction, CreateEnemyAction) and the
// action/timer types themselves live in package types so gameplay code can
// depend on them without importing this package.
package battle

// Package battle implements a real-time, timer-based combat system.
//
// Combat here is not strictly turn-based. Every participant charges an action
// timer continuously; when a timer fills, that participant may act. This gives
// an "active time battle" feel where faster entities act more often.
//
// # Participants
//
// Any object implementing BattleEntity can join a battle - in the basic-game
// example the game's own Player and Enemy entities do. An entity exposes a
// charging ActionTimer, battle stats (HP, speed), and a way to select the
// action it wants to perform.
//
// # The manager
//
// BattleManager is the orchestrator. Each frame it charges every entity's
// timer and, for any entity that is ready, produces an Action. Player
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
// The battle vocabulary - BattleEntity, Action, ActionType, ActionTimer,
// EntityStats and the construction helpers (CreatePlayerAction,
// CreateEnemyAction) - is owned by this package. Game entities that want to
// fight implement battle.BattleEntity, so the dependency points from the game
// into the system rather than leaking battle concepts into the engine's shared
// types package.
//
// # A decoupled add-on
//
// This package is an optional engine add-on rather than part of the core: the
// engine never imports it. It depends only on package types (the shared
// interface layer) and takes everything else through a Config value passed to
// NewBattleManager - queue size, timer charge rate, effect duration, and a
// Logger. It reads no global engine configuration, so it can be reused across
// games or tested in isolation. A game wires it up like:
//
//	bm := battle.NewBattleManager(battle.Config{
//		ActionQueueSize:      config.Global.Battle.ActionQueueSize,
//		TimerChargeRate:      config.Global.Battle.TimerChargeRate,
//		DamageEffectDuration: config.Global.Battle.DamageEffectDuration,
//		Logger:               logger.Logger,
//	})
//
// Pass battle.Config{} to accept all defaults.
package battle

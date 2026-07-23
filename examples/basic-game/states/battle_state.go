package states

import (
	"fmt"

	"example.com/basic-game/game/entities"
	"example.com/basic-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/prefab"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/systems"
	"github.com/cstevenson98/gowasm-engine/pkg/systems/battle"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BattleState is a turn-based (ATB) battle: background, player and enemy
// sprites (ECS entities), a battle menu, and the battle manager driving combat.
// Combat state lives in battle.Participant adapters (chunk-1 shim); a later pass
// folds it into components.
type BattleState struct {
	*state.BaseState

	playerPart *entities.Participant
	enemyPart  *entities.Participant

	menuSystem    *BattleMenuSystem
	battleManager *battle.BattleManager
	effectManager *battle.EffectManager
}

// NewBattleState creates the battle state.
func NewBattleState() *BattleState {
	return &BattleState{BaseState: state.NewBaseState("Battle")}
}

// Enter builds the battle scene and starts the battle system.
func (s *BattleState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	w := s.World()

	prefab.NewBackground(
		w,
		types.Vector2{X: 0, Y: 0},
		types.Vector2{X: s.ScreenWidth(), Y: s.ScreenHeight()},
		"assets/art/test-background.png",
	)

	// Player sprite (left) and enemy sprite (right) as static, animated entities.
	playerPos := types.Vector2{X: s.ScreenWidth() * 0.2, Y: s.ScreenHeight() * 0.5}
	enemyPos := types.Vector2{X: s.ScreenWidth() * 0.8, Y: s.ScreenHeight() * 0.5}

	entities.SpawnCharacter(w, playerPos,
		types.Vector2{X: gameconfig.Global.Player.Size, Y: gameconfig.Global.Player.Size},
		gameconfig.Global.Player.TexturePath, gameconfig.Global.Player.SpriteColumns, gameconfig.Global.Player.SpriteRows,
		s.DefaultFrameTime())
	entities.SpawnCharacter(w, enemyPos,
		types.Vector2{X: 32, Y: 64}, gameconfig.Global.Battle.EnemyTexture, 3, 2, s.DefaultFrameTime())

	// Combat participants (adapters for the battle manager).
	s.playerPart = entities.NewParticipant("Player", gameconfig.Global.Battle.PlayerHP, gameconfig.Global.Battle.PlayerMaxHP, playerPos, true)
	s.enemyPart = entities.NewParticipant("Enemy", gameconfig.Global.Battle.EnemyHP, gameconfig.Global.Battle.EnemyMaxHP, enemyPos, false)

	s.menuSystem = NewBattleMenuSystem()
	s.menuSystem.Initialize()
	s.menuSystem.SetActionCallback(s.enqueuePlayerAction)
	s.menuSystem.SetPlayer(s.playerPart)

	s.battleManager = battle.NewBattleManager(battle.Config{
		ActionQueueSize:      gameconfig.Global.Battle.ActionQueueSize,
		TimerChargeRate:      gameconfig.Global.Battle.TimerChargeRate,
		DamageEffectDuration: gameconfig.Global.Battle.DamageEffectDuration,
		Logger:               logger.Logger,
	})
	s.battleManager.AddEntity(s.playerPart)
	s.battleManager.AddEntity(s.enemyPart)
	s.effectManager = s.battleManager.GetEffectManager()

	// Animate the character sprites.
	s.Schedule().Add(systems.NewAnimation(w))

	return nil
}

// Update drives the battle system, effects, menu, and state switching.
func (s *BattleState) Update(dt float64) {
	s.BaseState.Update(dt) // animation + debug console

	if s.battleManager != nil {
		s.battleManager.Update(dt)
	}
	if s.effectManager != nil {
		s.effectManager.Update(dt)
	}
	if s.menuSystem != nil {
		s.menuSystem.Update(s.Input())
	}

	in := s.Input()
	if in.Key1Pressed && !in.Key1PressedLastFrame {
		if err := s.RequestState(types.GAMEPLAY); err != nil {
			logger.Logger.Errorf("Failed to switch to gameplay: %s", err.Error())
		}
	}
}

// Exit tears down the battle system and world.
func (s *BattleState) Exit() {
	s.battleManager = nil
	s.effectManager = nil
	s.playerPart = nil
	s.enemyPart = nil
	s.menuSystem = nil
	s.BaseState.Exit()
}

func (s *BattleState) enqueuePlayerAction(actionType battle.ActionType) {
	if s.battleManager == nil || s.playerPart == nil || s.enemyPart == nil {
		return
	}
	if action := battle.CreatePlayerAction(actionType, s.playerPart, s.enemyPart); action != nil {
		s.battleManager.EnqueueAction(action)
	}
}

// DrawOverlays renders the battle UI, effects, timer bars, and debug console.
func (s *BattleState) DrawOverlays() error {
	s.renderBattleMenu()
	s.renderDamageEffects()
	s.renderActionTimerBars()
	return s.BaseState.DrawOverlays()
}

func (s *BattleState) renderDamageEffects() {
	if s.effectManager == nil {
		return
	}
	ui := s.UI()
	for _, effect := range s.effectManager.GetActiveEffects() {
		alpha := effect.GetAlpha()
		color := types.Color{1, 0, 0, alpha}
		sign := "-"
		if effect.IsHealingEffect() {
			color = types.Color{0, 1, 0, alpha}
			sign = "+"
		}
		pos := effect.GetPosition()
		ui.TextColored(pos.X, pos.Y, color, fmt.Sprintf("%s%d", sign, effect.GetValue()))
	}
}

func (s *BattleState) renderActionTimerBars() {
	ui := s.UI()
	lineHeight := ui.LineHeight()
	if s.playerPart != nil {
		s.renderEntityTimerBar(s.playerPart, types.Vector2{X: 20, Y: 500}, "Player")
	}
	if s.enemyPart != nil {
		s.renderEntityTimerBar(s.enemyPart, types.Vector2{X: 20, Y: 500 + lineHeight}, "Enemy")
	}
}

func (s *BattleState) renderEntityTimerBar(entity battle.BattleEntity, position types.Vector2, label string) {
	current := entity.GetActionTimer().Current
	bar := "["
	segments := int(current / 0.2)
	for i := 0; i < 5; i++ {
		if i < segments {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	color := types.White
	if current >= 1.0 {
		color = types.Green
	}
	s.UI().TextColored(position.X, position.Y, color, fmt.Sprintf("%s: %s", label, bar))
}

func (s *BattleState) renderBattleMenu() {
	if s.menuSystem == nil {
		return
	}
	ui := s.UI()
	lineHeight := ui.LineHeight()

	if bl := s.menuSystem.battleLog; bl != nil {
		y := bl.position.Y
		for i, message := range bl.messages {
			if i >= bl.maxLines {
				break
			}
			ui.TextColored(bl.position.X, y, types.White, message)
			y += lineHeight
		}
	}

	if cs := s.menuSystem.characterStatus; cs != nil {
		ui.TextColored(cs.position.X, cs.position.Y, types.Green,
			fmt.Sprintf("Player: %d/%d HP", cs.playerHP, cs.playerMaxHP))
		ui.TextColored(cs.position.X, cs.position.Y+lineHeight, types.Red,
			fmt.Sprintf("Enemy: %d/%d HP", cs.enemyHP, cs.enemyMaxHP))
	}

	if am := s.menuSystem.actionMenu; am != nil {
		for i, action := range am.actions {
			text := "  " + action
			if i == am.selectedIndex {
				text = "> " + action
			}
			ui.TextColored(am.position.X, am.position.Y+float64(i)*lineHeight, types.Yellow, text)
		}
	}
}

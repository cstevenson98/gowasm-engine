package states

import (
	"example.com/basic-game/game/gameconfig"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/systems/battle"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BattleMenuSystem manages the battle UI: battle log, character status, and the
// action menu. It is plain immediate-mode UI state (no ECS entities).
type BattleMenuSystem struct {
	battleLog       *BattleLog
	characterStatus *CharacterStatus
	actionMenu      *ActionMenu

	onActionSelected func(battle.ActionType)
	player           battle.BattleEntity
}

// BattleLog displays battle messages.
type BattleLog struct {
	messages []string
	maxLines int
	position types.Vector2
}

// CharacterStatus displays player and enemy stats.
type CharacterStatus struct {
	playerHP, playerMaxHP int
	enemyHP, enemyMaxHP   int
	position              types.Vector2
}

// ActionMenu displays available actions with a selection indicator.
type ActionMenu struct {
	actions       []string
	selectedIndex int
	position      types.Vector2
}

// NewBattleMenuSystem creates the battle menu system.
func NewBattleMenuSystem() *BattleMenuSystem {
	return &BattleMenuSystem{}
}

// Initialize sets up the battle menu components.
func (bms *BattleMenuSystem) Initialize() {
	bms.battleLog = &BattleLog{
		messages: []string{"Battle begins!", "Player's turn"},
		maxLines: 8,
		position: types.Vector2{X: 20, Y: 20},
	}
	bms.characterStatus = &CharacterStatus{
		playerHP:    gameconfig.Global.Battle.PlayerHP,
		playerMaxHP: gameconfig.Global.Battle.PlayerMaxHP,
		enemyHP:     gameconfig.Global.Battle.EnemyHP,
		enemyMaxHP:  gameconfig.Global.Battle.EnemyMaxHP,
		position:    types.Vector2{X: 20, Y: 240},
	}
	bms.actionMenu = &ActionMenu{
		actions:  []string{"Attack", "Defend", "Item", "Run"},
		position: types.Vector2{X: 20, Y: 360},
	}
}

// Update handles battle menu navigation and action selection.
func (bms *BattleMenuSystem) Update(in types.InputState) {
	if bms.actionMenu == nil {
		return
	}

	if in.UpPressed && !in.UpPressedLastFrame {
		bms.actionMenu.selectedIndex = (bms.actionMenu.selectedIndex - 1 + len(bms.actionMenu.actions)) % len(bms.actionMenu.actions)
	}
	if in.DownPressed && !in.DownPressedLastFrame {
		bms.actionMenu.selectedIndex = (bms.actionMenu.selectedIndex + 1) % len(bms.actionMenu.actions)
	}

	if in.EnterPressed && !in.EnterPressedLastFrame {
		if bms.player != nil && !bms.player.IsReady() {
			bms.battleLog.AddMessage("Not ready yet! Wait for timer to fill.")
			return
		}
		selected := bms.actionMenu.actions[bms.actionMenu.selectedIndex]
		bms.battleLog.AddMessage("Selected: " + selected)
		actionType := convertStringToActionType(selected)
		if actionType != battle.ActionRun && bms.onActionSelected != nil {
			bms.onActionSelected(actionType)
		}
	}
}

// SetActionCallback registers the callback invoked when an action is chosen.
func (bms *BattleMenuSystem) SetActionCallback(cb func(battle.ActionType)) {
	bms.onActionSelected = cb
}

// SetPlayer sets the player reference used for readiness checks.
func (bms *BattleMenuSystem) SetPlayer(player battle.BattleEntity) {
	bms.player = player
}

func convertStringToActionType(action string) battle.ActionType {
	switch action {
	case "Attack":
		return battle.ActionAttack
	case "Defend":
		return battle.ActionDefend
	case "Item":
		return battle.ActionItem
	case "Run":
		return battle.ActionRun
	default:
		logger.Logger.Warnf("Unknown battle action: %s", action)
		return battle.ActionAttack
	}
}

// AddMessage appends a message to the battle log, capped at maxLines.
func (bl *BattleLog) AddMessage(message string) {
	bl.messages = append(bl.messages, message)
	if len(bl.messages) > bl.maxLines {
		bl.messages = bl.messages[len(bl.messages)-bl.maxLines:]
	}
}

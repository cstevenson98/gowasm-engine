//go:build js

package scenes

import (
	"github.com/conor/webgpu-triangle/pkg/canvas"
	"github.com/conor/webgpu-triangle/pkg/config"
	"github.com/conor/webgpu-triangle/pkg/logger"
	"github.com/conor/webgpu-triangle/pkg/text"
	"github.com/conor/webgpu-triangle/pkg/types"
)

// BattleMenuSystem manages the battle UI including battle log, character status, and action menu
type BattleMenuSystem struct {
	screenWidth  float64
	screenHeight float64

	// Menu components
	battleLog       *BattleLog
	characterStatus *CharacterStatus
	actionMenu      *ActionMenu

	// Text rendering
	textRenderer  text.TextRenderer
	font          text.Font
	canvasManager canvas.CanvasManager

	// Action callback
	onActionSelected func(types.ActionType)

	// Player reference for timer checking
	player types.BattleEntity
}

// BattleLog displays battle messages
type BattleLog struct {
	messages []string
	maxLines int
	position types.Vector2
	size     types.Vector2
}

// CharacterStatus displays player and enemy stats
type CharacterStatus struct {
	playerHP    int
	playerMaxHP int
	enemyHP     int
	enemyMaxHP  int
	position    types.Vector2
	size        types.Vector2
}

// ActionMenu displays available actions with selection indicator
type ActionMenu struct {
	actions       []string
	selectedIndex int
	position      types.Vector2
	size          types.Vector2
}

// NewBattleMenuSystem creates a new battle menu system
func NewBattleMenuSystem(screenWidth, screenHeight float64) *BattleMenuSystem {
	return &BattleMenuSystem{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

// Initialize sets up the battle menu system
func (bms *BattleMenuSystem) Initialize() {
	logger.Logger.Debugf("Initializing battle menu system")

	// Initialize battle log
	bms.battleLog = &BattleLog{
		messages: []string{
			"Battle begins!",
			"Player's turn",
		},
		maxLines: 8,
		position: types.Vector2{X: 20, Y: 20},
		size:     types.Vector2{X: 300, Y: 200},
	}

	// Initialize character status
	bms.characterStatus = &CharacterStatus{
		playerHP:    config.Global.Battle.PlayerHP,
		playerMaxHP: config.Global.Battle.PlayerMaxHP,
		enemyHP:     config.Global.Battle.EnemyHP,
		enemyMaxHP:  config.Global.Battle.EnemyMaxHP,
		position:    types.Vector2{X: 20, Y: 240},
		size:        types.Vector2{X: 300, Y: 100},
	}

	// Initialize action menu
	bms.actionMenu = &ActionMenu{
		actions: []string{
			"Attack",
			"Defend",
			"Item",
			"Run",
		},
		selectedIndex: 0,
		position:      types.Vector2{X: 20, Y: 360},
		size:          types.Vector2{X: 200, Y: 200},
	}

	logger.Logger.Debugf("Battle menu system initialized")
}

// Update updates the battle menu system
func (bms *BattleMenuSystem) Update(deltaTime float64, inputCapturer types.InputCapturer) {
	if inputCapturer == nil {
		return
	}

	inputState := inputCapturer.GetInputState()

	// Handle menu navigation
	if inputState.UpPressed && !inputState.UpPressedLastFrame {
		bms.actionMenu.selectedIndex--
		if bms.actionMenu.selectedIndex < 0 {
			bms.actionMenu.selectedIndex = len(bms.actionMenu.actions) - 1
		}
		logger.Logger.Debugf("Menu selection: %d", bms.actionMenu.selectedIndex)
	}

	if inputState.DownPressed && !inputState.DownPressedLastFrame {
		bms.actionMenu.selectedIndex++
		if bms.actionMenu.selectedIndex >= len(bms.actionMenu.actions) {
			bms.actionMenu.selectedIndex = 0
		}
		logger.Logger.Debugf("Menu selection: %d", bms.actionMenu.selectedIndex)
	}

	// Handle action selection (Enter key)
	if inputState.EnterPressed && !inputState.EnterPressedLastFrame {
		// Check if player is ready to act
		if bms.player != nil && !bms.player.IsReady() {
			bms.battleLog.AddMessage("Not ready yet! Wait for timer to fill.")
			logger.Logger.Debugf("Player not ready, timer at: %.2f", bms.player.GetActionTimer().Current)
			return
		}

		selectedAction := bms.actionMenu.actions[bms.actionMenu.selectedIndex]
		bms.battleLog.AddMessage("Selected: " + selectedAction)
		logger.Logger.Debugf("Action selected: %s", selectedAction)

		// Convert string action to ActionType and trigger callback
		actionType := bms.convertStringToActionType(selectedAction)
		if actionType != types.ActionRun && bms.onActionSelected != nil {
			bms.onActionSelected(actionType)
		}
	}
}

// GetRenderables returns all UI elements that need to be rendered
func (bms *BattleMenuSystem) GetRenderables() []types.GameObject {
	// For now, return empty - we'll implement text rendering later
	// This will be handled by the text rendering system
	return []types.GameObject{}
}

// Cleanup releases battle menu resources
func (bms *BattleMenuSystem) Cleanup() {
	logger.Logger.Debugf("Cleaning up battle menu system")
	bms.battleLog = nil
	bms.characterStatus = nil
	bms.actionMenu = nil
}

// AddMessage adds a message to the battle log
func (bl *BattleLog) AddMessage(message string) {
	bl.messages = append(bl.messages, message)

	// Keep only the last maxLines messages
	if len(bl.messages) > bl.maxLines {
		bl.messages = bl.messages[len(bl.messages)-bl.maxLines:]
	}
}

// GetMessages returns the current battle log messages
func (bl *BattleLog) GetMessages() []string {
	return bl.messages
}

// GetPosition returns the battle log position
func (bl *BattleLog) GetPosition() types.Vector2 {
	return bl.position
}

// GetSize returns the battle log size
func (bl *BattleLog) GetSize() types.Vector2 {
	return bl.size
}

// GetPlayerHP returns the player's current HP
func (cs *CharacterStatus) GetPlayerHP() int {
	return cs.playerHP
}

// GetPlayerMaxHP returns the player's maximum HP
func (cs *CharacterStatus) GetPlayerMaxHP() int {
	return cs.playerMaxHP
}

// GetEnemyHP returns the enemy's current HP
func (cs *CharacterStatus) GetEnemyHP() int {
	return cs.enemyHP
}

// GetEnemyMaxHP returns the enemy's maximum HP
func (cs *CharacterStatus) GetEnemyMaxHP() int {
	return cs.enemyMaxHP
}

// GetPosition returns the character status position
func (cs *CharacterStatus) GetPosition() types.Vector2 {
	return cs.position
}

// GetSize returns the character status size
func (cs *CharacterStatus) GetSize() types.Vector2 {
	return cs.size
}

// GetActions returns the available actions
func (am *ActionMenu) GetActions() []string {
	return am.actions
}

// GetSelectedIndex returns the currently selected action index
func (am *ActionMenu) GetSelectedIndex() int {
	return am.selectedIndex
}

// GetSelectedAction returns the currently selected action
func (am *ActionMenu) GetSelectedAction() string {
	if am.selectedIndex >= 0 && am.selectedIndex < len(am.actions) {
		return am.actions[am.selectedIndex]
	}
	return ""
}

// GetPosition returns the action menu position
func (am *ActionMenu) GetPosition() types.Vector2 {
	return am.position
}

// GetSize returns the action menu size
func (am *ActionMenu) GetSize() types.Vector2 {
	return am.size
}

// SetActionCallback sets the callback function for when an action is selected
func (bms *BattleMenuSystem) SetActionCallback(callback func(types.ActionType)) {
	bms.onActionSelected = callback
}

// SetPlayer sets the player reference for timer checking
func (bms *BattleMenuSystem) SetPlayer(player types.BattleEntity) {
	bms.player = player
}

// convertStringToActionType converts a string action to ActionType
func (bms *BattleMenuSystem) convertStringToActionType(action string) types.ActionType {
	switch action {
	case "Attack":
		return types.ActionAttack
	case "Defend":
		return types.ActionDefend
	case "Item":
		return types.ActionItem
	case "Run":
		return types.ActionRun
	default:
		return types.ActionAttack // Default fallback
	}
}

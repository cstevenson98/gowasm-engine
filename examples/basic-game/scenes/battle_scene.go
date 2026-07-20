package scenes

import (
	"fmt"

	"example.com/basic-game/game/entities"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/gameobject"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/systems/battle"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BattleScene represents a turn-based battle scene with player, enemy, and menu.
// It embeds BaseScene to inherit all common scene functionality.
type BattleScene struct {
	*pkscene.BaseScene

	// Battle participants
	player *entities.Player
	enemy  *entities.Enemy

	// Battle menu system
	menuSystem *BattleMenuSystem

	// Battle system
	battleManager *battle.BattleManager
	effectManager *battle.EffectManager
}

// NewBattleScene creates a new battle scene
func NewBattleScene(screenWidth, screenHeight float64) *BattleScene {
	baseScene := pkscene.NewBaseScene("Battle", screenWidth, screenHeight)

	// Declare required assets. Font textures are loaded automatically from
	// FontPaths, so they don't need to be listed under TexturePaths.
	baseScene.SetRequiredAssets(types.SceneAssets{
		TexturePaths: []string{
			"assets/art/test-background.png",
			config.Global.Player.TexturePath,
			config.Global.Battle.EnemyTexture,
		},
		FontPaths: []string{
			config.Global.Debug.FontPath,
		},
	})

	return &BattleScene{
		BaseScene: baseScene,
	}
}

// All interface implementations (SetInputCapturer, SetStateChangeCallback)
// are inherited from BaseScene

// RenderOverlays implements types.SceneOverlayRenderer by delegating to existing methods
func (s *BattleScene) RenderOverlays() error {
	if err := s.RenderBattleMenu(); err != nil {
		return err
	}
	if err := s.RenderDamageEffects(); err != nil {
		return err
	}
	if err := s.RenderActionTimerBars(); err != nil {
		return err
	}
	// Debug console rendered by BaseScene.RenderOverlays()
	return s.BaseScene.RenderOverlays()
}

// GetRequiredAssets is inherited from BaseScene (set in constructor)

// Initialize sets up the battle scene and creates game objects (overrides BaseScene.Initialize)
func (s *BattleScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.GetName())

	// Call base initialization (sets up layers + debug console)
	if err := s.BaseScene.Initialize(); err != nil {
		return err
	}

	// Create background (BACKGROUND layer)
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.GetScreenWidth(), Y: s.GetScreenHeight()},
		"assets/art/test-background.png",
	)
	s.AddGameObject(pkscene.BACKGROUND, background)
	logger.Logger.Debugf("Created Background in %s scene", s.GetName())

	// Create player on the left side (ENTITIES layer)
	playerX := s.GetScreenWidth() * 0.2  // 20% from left
	playerY := s.GetScreenHeight() * 0.5 // Center vertically
	s.player = entities.NewPlayer(
		types.Vector2{X: playerX, Y: playerY},
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)
	logger.Logger.Debugf("Created Player on left side in %s scene", s.GetName())

	// Create enemy on the right side (ENTITIES layer)
	enemyX := s.GetScreenWidth() * 0.8  // 80% from left (right side)
	enemyY := s.GetScreenHeight() * 0.5 // Center vertically
	s.enemy = entities.NewEnemy(
		types.Vector2{X: enemyX, Y: enemyY},
		types.Vector2{X: 32.0, Y: 64.0}, // Ghost sprite dimensions (96x128 total, 3x2 grid = 32x64 per frame)
		config.Global.Battle.EnemyTexture,
	)
	logger.Logger.Debugf("Created Enemy on right side in %s scene", s.GetName())

	// Initialize battle menu system
	s.menuSystem = NewBattleMenuSystem(s.GetScreenWidth(), s.GetScreenHeight())
	s.menuSystem.Initialize()

	// Set up action callback
	s.menuSystem.SetActionCallback(s.EnqueuePlayerAction)

	// Set player reference for timer checking
	s.menuSystem.SetPlayer(s.player)

	// Initialize battle system, wiring the engine's global config and logger
	// into the add-on's injected Config.
	s.battleManager = battle.NewBattleManager(battle.Config{
		ActionQueueSize:      config.Global.Battle.ActionQueueSize,
		TimerChargeRate:      config.Global.Battle.TimerChargeRate,
		DamageEffectDuration: config.Global.Battle.DamageEffectDuration,
		Logger:               logger.Logger,
	})

	// Add entities to battle manager
	s.battleManager.AddEntity(s.player)
	s.battleManager.AddEntity(s.enemy)

	// Get effect manager from battle manager
	s.effectManager = s.battleManager.GetEffectManager()

	// Start battle processing
	s.battleManager.StartProcessing()

	return nil
}

// Update updates all game objects in the scene
func (s *BattleScene) Update(deltaTime float64) {
	// Update battle system and visual effects.
	if s.battleManager != nil {
		s.battleManager.Update(deltaTime)
	}
	if s.effectManager != nil {
		s.effectManager.Update(deltaTime)
	}

	// Advance the battle participants. They're held as fields (not in layers),
	// so they're pumped explicitly; BaseGameObject.Update animates the sprite
	// (they don't move in battle).
	if s.player != nil {
		s.player.Update(deltaTime)
	}
	if s.enemy != nil {
		s.enemy.Update(deltaTime)
	}

	// Advance objects held in layers (the background).
	s.BaseScene.Update(deltaTime)

	// Battle menu input.
	if s.menuSystem != nil {
		s.menuSystem.Update(deltaTime, s.GetInputCapturer())
	}

	inputState := s.GetInputState()

	// Scene switching: Key 1 returns to gameplay. The engine defers the switch
	// to the next frame, so it is safe to continue this Update.
	if inputState.Key1Pressed && !inputState.Key1PressedLastFrame {
		logger.Logger.Debugf("Key 1 pressed: switching to gameplay scene")
		if err := s.RequestStateChange(types.GAMEPLAY); err != nil {
			logger.Logger.Errorf("Failed to switch to gameplay scene: %s", err.Error())
		}
	}
}

// RenderDamageEffects renders damage/healing numbers
func (s *BattleScene) RenderDamageEffects() error {
	if s.effectManager == nil {
		return nil
	}

	for _, effect := range s.effectManager.GetActiveEffects() {
		alpha := effect.GetAlpha()

		color := types.Color{1, 0, 0, alpha} // Red for damage
		sign := "-"
		if effect.IsHealingEffect() {
			color = types.Color{0, 1, 0, alpha} // Green for healing
			sign = "+"
		}

		pos := effect.GetPosition()
		s.UI().TextColored(pos.X, pos.Y, color, fmt.Sprintf("%s%d", sign, effect.GetValue()))
	}

	return nil
}

// RenderActionTimerBars renders action timer bars for player and enemy
func (s *BattleScene) RenderActionTimerBars() error {
	lineHeight := s.UI().LineHeight()

	if s.player != nil {
		s.renderEntityTimerBar(s.player, types.Vector2{X: 20, Y: 500}, "Player")
	}

	if s.enemy != nil {
		s.renderEntityTimerBar(s.enemy, types.Vector2{X: 20, Y: 500 + lineHeight}, "Enemy")
	}

	return nil
}

// renderEntityTimerBar renders a timer bar for a specific entity
func (s *BattleScene) renderEntityTimerBar(entity battle.BattleEntity, position types.Vector2, label string) {
	timer := entity.GetActionTimer()
	current := timer.Current

	// Create timer bar: [=====] format
	bar := "["

	// Add = characters based on timer progress
	if current >= 0.2 {
		bar += "="
	}
	if current >= 0.4 {
		bar += "="
	}
	if current >= 0.6 {
		bar += "="
	}
	if current >= 0.8 {
		bar += "="
	}
	if current >= 1.0 {
		bar += "="
	}

	// Add spaces for remaining segments
	segments := int(current / 0.2)
	for i := segments; i < 5; i++ {
		bar += " "
	}

	bar += "]"

	// Add label
	fullText := fmt.Sprintf("%s: %s", label, bar)

	// Determine color based on readiness
	color := types.White // White when charging
	if current >= 1.0 {
		color = types.Green // Green when ready
	}

	s.UI().TextColored(position.X, position.Y, color, fullText)
}

// RenderBattleMenu renders the battle menu UI
func (s *BattleScene) RenderBattleMenu() error {
	if s.menuSystem == nil {
		return nil
	}

	lineHeight := s.UI().LineHeight()

	// Render battle log
	if battleLog := s.menuSystem.battleLog; battleLog != nil {
		pos := battleLog.GetPosition()
		y := pos.Y
		for i, message := range battleLog.GetMessages() {
			if i >= battleLog.maxLines {
				break
			}
			s.UI().TextColored(pos.X, y, types.White, message)
			y += lineHeight
		}
	}

	// Render character status
	if characterStatus := s.menuSystem.characterStatus; characterStatus != nil {
		pos := characterStatus.GetPosition()
		s.UI().TextColored(pos.X, pos.Y, types.Green,
			fmt.Sprintf("Player: %d/%d HP", characterStatus.GetPlayerHP(), characterStatus.GetPlayerMaxHP()))
		s.UI().TextColored(pos.X, pos.Y+lineHeight, types.Red,
			fmt.Sprintf("Enemy: %d/%d HP", characterStatus.GetEnemyHP(), characterStatus.GetEnemyMaxHP()))
	}

	// Render action menu
	if actionMenu := s.menuSystem.actionMenu; actionMenu != nil {
		pos := actionMenu.GetPosition()
		for i, action := range actionMenu.GetActions() {
			displayText := "  " + action
			if i == actionMenu.GetSelectedIndex() {
				displayText = "> " + action
			}
			s.UI().TextColored(pos.X, pos.Y+float64(i)*lineHeight, types.Yellow, displayText)
		}
	}

	return nil
}

// GetRenderables returns all game objects in the correct render order
func (s *BattleScene) GetRenderables() []types.GameObject {
	var result []types.GameObject

	// Render layers in order: BACKGROUND → ENTITIES → UI
	for _, layer := range []pkscene.SceneLayer{pkscene.BACKGROUND, pkscene.ENTITIES, pkscene.UI} {
		// Add player to ENTITIES layer during rendering
		if layer == pkscene.ENTITIES && s.player != nil {
			result = append(result, s.player)
		}

		// Add enemy to ENTITIES layer during rendering
		if layer == pkscene.ENTITIES && s.enemy != nil {
			result = append(result, s.enemy)
		}

		// Add other game objects in this layer
		result = append(result, s.GetLayer(layer)...)
	}

	// Add battle menu UI elements
	if s.menuSystem != nil {
		menuRenderables := s.menuSystem.GetRenderables()
		result = append(result, menuRenderables...)
	}

	return result
}

// Cleanup releases scene resources (overrides BaseScene.Cleanup)
func (s *BattleScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.GetName())

	// Stop battle system
	if s.battleManager != nil {
		s.battleManager.StopProcessing()
		s.battleManager = nil
	}

	// Clear effect manager
	s.effectManager = nil

	// Clear player and enemy references
	s.player = nil
	s.enemy = nil

	// Cleanup menu system
	if s.menuSystem != nil {
		s.menuSystem.Cleanup()
		s.menuSystem = nil
	}

	// Call base cleanup (clears layers)
	s.BaseScene.Cleanup()
}

// GetName, AddGameObject, RemoveGameObject are inherited from BaseScene

// GetPlayer returns the player object
func (s *BattleScene) GetPlayer() *entities.Player {
	return s.player
}

// GetEnemy returns the enemy object
func (s *BattleScene) GetEnemy() *entities.Enemy {
	return s.enemy
}

// EnqueuePlayerAction creates and enqueues a player action
func (s *BattleScene) EnqueuePlayerAction(actionType battle.ActionType) {
	if s.battleManager == nil || s.player == nil || s.enemy == nil {
		return
	}

	// Create the action using the battle system
	action := battle.CreatePlayerAction(actionType, s.player, s.enemy)
	if action != nil {
		s.battleManager.EnqueueAction(action)
		logger.Logger.Debugf("Enqueued player action: %s", actionType.String())
	}
}

// GetBattleManager returns the battle manager (for external access)
func (s *BattleScene) GetBattleManager() *battle.BattleManager {
	return s.battleManager
}

// GetEffectManager returns the effect manager (for external access)
func (s *BattleScene) GetEffectManager() *battle.EffectManager {
	return s.effectManager
}

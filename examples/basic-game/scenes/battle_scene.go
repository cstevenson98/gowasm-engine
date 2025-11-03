//go:build js

package scenes

import (
	"fmt"

	"github.com/cstevenson98/gowasm-engine/pkg/battle"
	"github.com/cstevenson98/gowasm-engine/pkg/canvas"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/debug"
	"github.com/cstevenson98/gowasm-engine/pkg/gameobject"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	pkscene "github.com/cstevenson98/gowasm-engine/pkg/scene"
	"github.com/cstevenson98/gowasm-engine/pkg/text"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BattleScene represents a turn-based battle scene with player, enemy, and menu
type BattleScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// State change callback (injected by engine)
	stateChangeCallback func(state types.PipelineState) error

	// Battle participants
	player *gameobject.Player
	enemy  *gameobject.Enemy

	// Battle menu system
	menuSystem *BattleMenuSystem

	// Battle system
	battleManager *battle.BattleManager
	effectManager *battle.EffectManager

	// Game objects organized by layer
	layers map[pkscene.SceneLayer][]types.GameObject

	// Debug rendering
	debugFont         text.Font
	debugTextRenderer text.TextRenderer
	canvasManager     canvas.CanvasManager

	// Menu text rendering
	menuFont         text.Font
	menuTextRenderer text.TextRenderer

	// Debug console toggle state
	f2PressedLastFrame bool

	// Key press state tracking
	key1PressedLastFrame bool
	key2PressedLastFrame bool
}

// NewBattleScene creates a new battle scene
func NewBattleScene(screenWidth, screenHeight float64) *BattleScene {
	return &BattleScene{
		name:         "Battle",
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		layers:       make(map[pkscene.SceneLayer][]types.GameObject),
	}
}

// SetInputCapturer implements types.SceneInputProvider
func (s *BattleScene) SetInputCapturer(inputCapturer types.InputCapturer) {
	s.inputCapturer = inputCapturer
}

// SetStateChangeCallback implements types.SceneChangeRequester
func (s *BattleScene) SetStateChangeCallback(callback func(state types.PipelineState) error) {
	s.stateChangeCallback = callback
}

// SetCanvasManager sets the canvas manager for debug rendering
func (s *BattleScene) SetCanvasManager(cm canvas.CanvasManager) {
	s.canvasManager = cm
}

// InitializeDebugConsole initializes the debug console font and text renderer
func (s *BattleScene) InitializeDebugConsole() error {
	if !config.Global.Debug.Enabled {
		return nil
	}

	logger.Logger.Debugf("Initializing debug console for %s scene", s.name)

	// Create and load font metadata
	s.debugFont = text.NewSpriteFont()
	err := s.debugFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load debug font: %s", err)
		return err
	}

	// Create text renderer (texture will be loaded by engine's loadSpriteTextures)
	s.debugTextRenderer = text.NewTextRenderer(s.canvasManager)

	logger.Logger.Debugf("Debug console initialized successfully")
	// Post a welcome message after a short delay to allow texture loading
	debug.Console.PostMessage("System", "Battle scene ready")

	return nil
}

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
	if err := s.RenderDebugConsole(); err != nil {
		return err
	}
	return nil
}

// GetExtraTexturePaths implements types.SceneTextureProvider
func (s *BattleScene) GetExtraTexturePaths() []string {
	var paths []string
	if s.debugFont != nil && s.debugFont.IsLoaded() {
		paths = append(paths, s.debugFont.GetTexturePath())
	}
	if s.menuFont != nil && s.menuFont.IsLoaded() {
		paths = append(paths, s.menuFont.GetTexturePath())
	}
	return paths
}

// GetRequiredAssets implements types.SceneAssetProvider
func (s *BattleScene) GetRequiredAssets() types.SceneAssets {
	return types.SceneAssets{
		TexturePaths: []string{
			"art/test-background.png",
			config.Global.Player.TexturePath,
			config.Global.Battle.EnemyTexture,
		},
		FontPaths: []string{
			config.Global.Debug.FontPath, // Same font used for debug and menu, cached once
		},
	}
}

// InitializeMenuText initializes the menu text rendering system
func (s *BattleScene) InitializeMenuText() error {
	logger.Logger.Debugf("Initializing menu text rendering for %s scene", s.name)

	// Create and load font metadata for menu text
	s.menuFont = text.NewSpriteFont()
	err := s.menuFont.(*text.SpriteFont).LoadFont(config.Global.Debug.FontPath)
	if err != nil {
		logger.Logger.Errorf("Failed to load menu font: %s", err)
		return err
	}

	// Create text renderer for menu
	s.menuTextRenderer = text.NewTextRenderer(s.canvasManager)

	logger.Logger.Debugf("Menu text rendering initialized successfully")
	return nil
}

// Initialize sets up the battle scene and creates game objects
func (s *BattleScene) Initialize() error {
	logger.Logger.Debugf("Initializing %s scene", s.name)

	// Initialize layer slices
	s.layers[pkscene.BACKGROUND] = []types.GameObject{}
	s.layers[pkscene.ENTITIES] = []types.GameObject{}
	s.layers[pkscene.UI] = []types.GameObject{}

	// Create background (BACKGROUND layer)
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.screenWidth, Y: s.screenHeight},
		"art/test-background.png",
	)
	s.AddGameObject(pkscene.BACKGROUND, background)
	logger.Logger.Debugf("Created Background in %s scene", s.name)

	// Create player on the left side (ENTITIES layer)
	playerX := s.screenWidth * 0.2  // 20% from left
	playerY := s.screenHeight * 0.5 // Center vertically
	s.player = gameobject.NewPlayer(
		types.Vector2{X: playerX, Y: playerY},
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
	)
	logger.Logger.Debugf("Created Player on left side in %s scene", s.name)

	// Create enemy on the right side (ENTITIES layer)
	enemyX := s.screenWidth * 0.8  // 80% from left (right side)
	enemyY := s.screenHeight * 0.5 // Center vertically
	s.enemy = gameobject.NewEnemy(
		types.Vector2{X: enemyX, Y: enemyY},
		types.Vector2{X: 32.0, Y: 64.0}, // Ghost sprite dimensions (96x128 total, 3x2 grid = 32x64 per frame)
		config.Global.Battle.EnemyTexture,
	)
	logger.Logger.Debugf("Created Enemy on right side in %s scene", s.name)

	// Initialize battle menu system
	s.menuSystem = NewBattleMenuSystem(s.screenWidth, s.screenHeight)
	s.menuSystem.Initialize()

	// Set up action callback
	s.menuSystem.SetActionCallback(s.EnqueuePlayerAction)

	// Set player reference for timer checking
	s.menuSystem.SetPlayer(s.player)

	// Initialize battle system
	s.battleManager = battle.NewBattleManager()

	// Add entities to battle manager
	s.battleManager.AddEntity(s.player)
	s.battleManager.AddEntity(s.enemy)

	// Get effect manager from battle manager
	s.effectManager = s.battleManager.GetEffectManager()

	// Start battle processing
	s.battleManager.StartProcessing()

	// Initialize menu text rendering
	err := s.InitializeMenuText()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize menu text: %s", err)
	}

	return nil
}

// Update updates all game objects in the scene
func (s *BattleScene) Update(deltaTime float64) {
	// Update battle system
	if s.battleManager != nil {
		s.battleManager.Update(deltaTime)
	}

	// Update effect manager
	if s.effectManager != nil {
		s.effectManager.Update(deltaTime)
	}

	// Update player (no input handling in battle - menu handles input)
	if s.player != nil {
		// Update player sprite (animation only, no movement)
		if sprite := s.player.GetSprite(); sprite != nil {
			sprite.Update(deltaTime)
		}
		s.player.Update(deltaTime)
	}

	// Update enemy (animation only)
	if s.enemy != nil {
		if sprite := s.enemy.GetSprite(); sprite != nil {
			sprite.Update(deltaTime)
		}
		s.enemy.Update(deltaTime)
	}

	// Update all game objects in all layers
	for _, layer := range []pkscene.SceneLayer{pkscene.BACKGROUND, pkscene.ENTITIES, pkscene.UI} {
		for _, gameObject := range s.layers[layer] {
			if mover := gameObject.GetMover(); mover != nil {
				mover.Update(deltaTime)
			}

			if sprite := gameObject.GetSprite(); sprite != nil {
				sprite.Update(deltaTime)
			}

			gameObject.Update(deltaTime)
		}
	}

	// Update battle menu system
	if s.menuSystem != nil {
		s.menuSystem.Update(deltaTime, s.inputCapturer)
	}

	// Handle debug console toggle (F2) and scene switching (Key 1)
	if s.inputCapturer != nil {
		inputState := s.inputCapturer.GetInputState()
		// Debug logging to see what keys are being pressed
		if inputState.F2Pressed {
			logger.Logger.Debugf("F2 key detected: %t, LastFrame: %t", inputState.F2Pressed, s.f2PressedLastFrame)
		}
		// Check for F2 key press using local state
		if inputState.F2Pressed && !s.f2PressedLastFrame {
			debug.Console.ToggleVisibility()
			logger.Logger.Debugf("Debug console toggled via F2")
		}
		// Update local state
		s.f2PressedLastFrame = inputState.F2Pressed

		// Handle scene switching: Key 1 switches to gameplay scene, Key 2 to battle (no-op, already in battle)
		if inputState.Key1Pressed && !s.key1PressedLastFrame && s.stateChangeCallback != nil {
			logger.Logger.Debugf("Key 1 pressed: switching to gameplay scene")
			err := s.stateChangeCallback(types.GAMEPLAY)
			if err != nil {
				logger.Logger.Errorf("Failed to switch to gameplay scene: %s", err.Error())
			}
			// Return early - scene may have been cleaned up during state change
			s.key1PressedLastFrame = inputState.Key1Pressed
			s.key2PressedLastFrame = inputState.Key2Pressed
			return
		}
		s.key1PressedLastFrame = inputState.Key1Pressed
		s.key2PressedLastFrame = inputState.Key2Pressed
	}

	// Update debug console
	if config.Global.Debug.Enabled {
		debug.Console.Update(deltaTime)
	}
}

// RenderDebugConsole renders the debug console UI
func (s *BattleScene) RenderDebugConsole() error {
	if !config.Global.Debug.Enabled || s.debugFont == nil || s.debugTextRenderer == nil {
		return nil
	}

	return debug.Console.Render(s.canvasManager, s.debugTextRenderer, s.debugFont)
}

// RenderDamageEffects renders damage/healing numbers
func (s *BattleScene) RenderDamageEffects() error {
	if s.effectManager == nil || s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	effects := s.effectManager.GetActiveEffects()
	for _, effect := range effects {
		pos := effect.GetPosition()
		value := effect.GetValue()
		alpha := effect.GetAlpha()

		// Determine color based on effect type
		var color [4]float32
		if effect.IsHealingEffect() {
			color = [4]float32{0.0, 1.0, 0.0, alpha} // Green for healing
		} else {
			color = [4]float32{1.0, 0.0, 0.0, alpha} // Red for damage
		}

		// Format the damage/healing text
		var text string
		if effect.IsHealingEffect() {
			text = fmt.Sprintf("+%d", value)
		} else {
			text = fmt.Sprintf("-%d", value)
		}

		// Render the text
		err := s.menuTextRenderer.RenderText(
			text,
			pos,
			s.menuFont,
			color,
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render damage effect: %s", err)
		}
	}

	return nil
}

// RenderActionTimerBars renders action timer bars for player and enemy
func (s *BattleScene) RenderActionTimerBars() error {
	if s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	// Calculate line height accounting for pixel scale
	_, cellHeight := s.menuFont.GetCellSize()
	lineHeight := float64(cellHeight)
	if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
		lineHeight *= float64(config.Global.Rendering.PixelScale)
	}
	lineHeight *= config.Global.Rendering.UILineSpacing // UI line spacing

	// Render player timer bar
	if s.player != nil {
		s.renderEntityTimerBar(s.player, types.Vector2{X: 20, Y: 500}, "Player")
	}

	// Render enemy timer bar
	if s.enemy != nil {
		s.renderEntityTimerBar(s.enemy, types.Vector2{X: 20, Y: 500 + lineHeight}, "Enemy")
	}

	return nil
}

// renderEntityTimerBar renders a timer bar for a specific entity
func (s *BattleScene) renderEntityTimerBar(entity types.BattleEntity, position types.Vector2, label string) {
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
	var color [4]float32
	if current >= 1.0 {
		color = [4]float32{0.0, 1.0, 0.0, 1.0} // Green when ready
	} else {
		color = [4]float32{1.0, 1.0, 1.0, 1.0} // White when charging
	}

	// Render the timer bar
	err := s.menuTextRenderer.RenderText(
		fullText,
		position,
		s.menuFont,
		color,
	)
	if err != nil {
		logger.Logger.Tracef("Failed to render timer bar: %s", err)
	}
}

// RenderBattleMenu renders the battle menu UI
func (s *BattleScene) RenderBattleMenu() error {
	if s.menuSystem == nil || s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	// Render battle log
	battleLog := s.menuSystem.battleLog
	if battleLog != nil {
		// Calculate line height accounting for pixel scale
		_, cellHeight := s.menuFont.GetCellSize()
		lineHeight := float64(cellHeight)
		if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
			lineHeight *= float64(config.Global.Rendering.PixelScale)
		}
		lineHeight *= config.Global.Rendering.UILineSpacing // UI line spacing

		y := battleLog.GetPosition().Y
		for i, message := range battleLog.GetMessages() {
			if i >= battleLog.maxLines {
				break
			}
			err := s.menuTextRenderer.RenderText(
				message,
				types.Vector2{X: battleLog.GetPosition().X, Y: y},
				s.menuFont,
				[4]float32{1.0, 1.0, 1.0, 1.0}, // White text
			)
			if err != nil {
				logger.Logger.Tracef("Failed to render battle log message: %s", err)
			}
			y += lineHeight // Line spacing with pixel scale
		}
	}

	// Render character status
	characterStatus := s.menuSystem.characterStatus
	if characterStatus != nil {
		pos := characterStatus.GetPosition()

		// Calculate line height accounting for pixel scale
		_, cellHeight := s.menuFont.GetCellSize()
		lineHeight := float64(cellHeight)
		if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
			lineHeight *= float64(config.Global.Rendering.PixelScale)
		}
		lineHeight *= config.Global.Rendering.UILineSpacing // UI line spacing

		// Player status
		playerText := fmt.Sprintf("Player: %d/%d HP", characterStatus.GetPlayerHP(), characterStatus.GetPlayerMaxHP())
		err := s.menuTextRenderer.RenderText(
			playerText,
			types.Vector2{X: pos.X, Y: pos.Y},
			s.menuFont,
			[4]float32{0.0, 1.0, 0.0, 1.0}, // Green text for player
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render player status: %s", err)
		}

		// Enemy status
		enemyText := fmt.Sprintf("Enemy: %d/%d HP", characterStatus.GetEnemyHP(), characterStatus.GetEnemyMaxHP())
		err = s.menuTextRenderer.RenderText(
			enemyText,
			types.Vector2{X: pos.X, Y: pos.Y + lineHeight},
			s.menuFont,
			[4]float32{1.0, 0.0, 0.0, 1.0}, // Red text for enemy
		)
		if err != nil {
			logger.Logger.Tracef("Failed to render enemy status: %s", err)
		}
	}

	// Render action menu
	actionMenu := s.menuSystem.actionMenu
	if actionMenu != nil {
		pos := actionMenu.GetPosition()
		actions := actionMenu.GetActions()
		selectedIndex := actionMenu.GetSelectedIndex()

		// Calculate line height accounting for pixel scale
		_, cellHeight := s.menuFont.GetCellSize()
		lineHeight := float64(cellHeight)
		if config.Global.Rendering.PixelPerfectScaling && config.Global.Rendering.PixelScale > 1 {
			lineHeight *= float64(config.Global.Rendering.PixelScale)
		}
		lineHeight *= config.Global.Rendering.UILineSpacing // UI line spacing

		for i, action := range actions {
			// Add selection indicator
			displayText := action
			if i == selectedIndex {
				displayText = "> " + action
			} else {
				displayText = "  " + action
			}

			err := s.menuTextRenderer.RenderText(
				displayText,
				types.Vector2{X: pos.X, Y: pos.Y + float64(i)*lineHeight},
				s.menuFont,
				[4]float32{1.0, 1.0, 0.0, 1.0}, // Yellow text for menu
			)
			if err != nil {
				logger.Logger.Tracef("Failed to render action menu item: %s", err)
			}
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
		result = append(result, s.layers[layer]...)
	}

	// Add battle menu UI elements
	if s.menuSystem != nil {
		menuRenderables := s.menuSystem.GetRenderables()
		result = append(result, menuRenderables...)
	}

	return result
}

// Cleanup releases scene resources
func (s *BattleScene) Cleanup() {
	logger.Logger.Debugf("Cleaning up %s scene", s.name)

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

	// Clear all layers
	for layer := range s.layers {
		s.layers[layer] = nil
	}
	s.layers = make(map[pkscene.SceneLayer][]types.GameObject)
}

// GetName returns the scene identifier
func (s *BattleScene) GetName() string {
	return s.name
}

// AddGameObject adds a game object to the specified layer
func (s *BattleScene) AddGameObject(layer pkscene.SceneLayer, obj types.GameObject) {
	s.layers[layer] = append(s.layers[layer], obj)
	logger.Logger.Debugf("Added GameObject to %s layer in %s scene", layer.String(), s.name)
}

// RemoveGameObject removes a game object from the specified layer
func (s *BattleScene) RemoveGameObject(layer pkscene.SceneLayer, obj types.GameObject) {
	objects := s.layers[layer]
	for i, o := range objects {
		if o == obj {
			s.layers[layer] = append(objects[:i], objects[i+1:]...)
			logger.Logger.Debugf("Removed GameObject from %s layer in %s scene", layer.String(), s.name)
			return
		}
	}
}

// GetPlayer returns the player object
func (s *BattleScene) GetPlayer() *gameobject.Player {
	return s.player
}

// GetEnemy returns the enemy object
func (s *BattleScene) GetEnemy() *gameobject.Enemy {
	return s.enemy
}

// GetDebugFont returns the debug font (for texture loading)
func (s *BattleScene) GetDebugFont() text.Font {
	return s.debugFont
}

// GetMenuFont returns the menu font (for texture loading)
func (s *BattleScene) GetMenuFont() text.Font {
	return s.menuFont
}

// EnqueuePlayerAction creates and enqueues a player action
func (s *BattleScene) EnqueuePlayerAction(actionType types.ActionType) {
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

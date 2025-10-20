//go:build js

package scene

import (
	"fmt"

	"github.com/conor/webgpu-triangle/internal/canvas"
	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/debug"
	"github.com/conor/webgpu-triangle/internal/gameobject"
	"github.com/conor/webgpu-triangle/internal/logger"
	"github.com/conor/webgpu-triangle/internal/text"
	"github.com/conor/webgpu-triangle/internal/types"
)

// BattleScene represents a turn-based battle scene with player, enemy, and menu
type BattleScene struct {
	name          string
	screenWidth   float64
	screenHeight  float64
	inputCapturer types.InputCapturer

	// Battle participants
	player *gameobject.Player
	enemy  *gameobject.Enemy

	// Battle menu system
	menuSystem *BattleMenuSystem

	// Game objects organized by layer
	layers map[SceneLayer][]types.GameObject

	// Debug rendering
	debugFont         text.Font
	debugTextRenderer text.TextRenderer
	canvasManager     canvas.CanvasManager

	// Menu text rendering
	menuFont         text.Font
	menuTextRenderer text.TextRenderer
	
	// Debug console toggle state
	f2PressedLastFrame bool
}

// NewBattleScene creates a new battle scene
func NewBattleScene(screenWidth, screenHeight float64, inputCapturer types.InputCapturer) *BattleScene {
	return &BattleScene{
		name:          "Battle",
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		inputCapturer: inputCapturer,
		layers:        make(map[SceneLayer][]types.GameObject),
	}
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
	s.layers[BACKGROUND] = []types.GameObject{}
	s.layers[ENTITIES] = []types.GameObject{}
	s.layers[UI] = []types.GameObject{}

	// Create background (BACKGROUND layer)
	background := gameobject.NewBackground(
		types.Vector2{X: 0, Y: 0}, // Top-left corner
		types.Vector2{X: s.screenWidth, Y: s.screenHeight},
		"art/test-background.png",
	)
	s.AddGameObject(BACKGROUND, background)
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
	enemyX := s.screenWidth * 0.8   // 80% from left (right side)
	enemyY := s.screenHeight * 0.5  // Center vertically
	s.enemy = gameobject.NewEnemy(
		types.Vector2{X: enemyX, Y: enemyY},
		types.Vector2{X: 32.0, Y: 64.0}, // Ghost sprite dimensions (96x128 total, 3x2 grid = 32x64 per frame)
		config.Global.Battle.EnemyTexture,
	)
	logger.Logger.Debugf("Created Enemy on right side in %s scene", s.name)

	// Initialize battle menu system
	s.menuSystem = NewBattleMenuSystem(s.screenWidth, s.screenHeight)
	s.menuSystem.Initialize()

	// Initialize menu text rendering
	err := s.InitializeMenuText()
	if err != nil {
		logger.Logger.Warnf("Failed to initialize menu text: %s", err)
	}

	return nil
}

// Update updates all game objects in the scene
func (s *BattleScene) Update(deltaTime float64) {
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
	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
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

	// Handle debug console toggle (F2)
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

// RenderBattleMenu renders the battle menu UI
func (s *BattleScene) RenderBattleMenu() error {
	if s.menuSystem == nil || s.menuFont == nil || s.menuTextRenderer == nil {
		return nil
	}

	// Render battle log
	battleLog := s.menuSystem.battleLog
	if battleLog != nil {
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
			y += 20 // Line spacing
		}
	}

	// Render character status
	characterStatus := s.menuSystem.characterStatus
	if characterStatus != nil {
		pos := characterStatus.GetPosition()
		
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
			types.Vector2{X: pos.X, Y: pos.Y + 20},
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
				types.Vector2{X: pos.X, Y: pos.Y + float64(i*25)},
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
	for _, layer := range []SceneLayer{BACKGROUND, ENTITIES, UI} {
		// Add player to ENTITIES layer during rendering
		if layer == ENTITIES && s.player != nil {
			result = append(result, s.player)
		}

		// Add enemy to ENTITIES layer during rendering
		if layer == ENTITIES && s.enemy != nil {
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
	s.layers = make(map[SceneLayer][]types.GameObject)
}

// GetName returns the scene identifier
func (s *BattleScene) GetName() string {
	return s.name
}

// AddGameObject adds a game object to the specified layer
func (s *BattleScene) AddGameObject(layer SceneLayer, obj types.GameObject) {
	s.layers[layer] = append(s.layers[layer], obj)
	logger.Logger.Debugf("Added GameObject to %s layer in %s scene", layer.String(), s.name)
}

// RemoveGameObject removes a game object from the specified layer
func (s *BattleScene) RemoveGameObject(layer SceneLayer, obj types.GameObject) {
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

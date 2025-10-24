//go:build js

package gameobject

import (
	"sync"

	"github.com/conor/webgpu-triangle/internal/config"
	"github.com/conor/webgpu-triangle/internal/mover"
	"github.com/conor/webgpu-triangle/internal/sprite"
	"github.com/conor/webgpu-triangle/internal/types"
)

// Player is a GameObject that represents the player character
type Player struct {
	sprite types.Sprite
	mover  types.Mover
	state  types.ObjectState

	moveSpeed float64 // Base movement speed in pixels per second

	// Debug message timing
	debugMessageTimer    float64
	debugMessageInterval float64 // Post debug message every N seconds

	// Battle system
	actionTimer    *types.ActionTimer
	stats          *types.EntityStats
	selectedAction types.ActionType // Player's selected action from menu

	mu sync.Mutex
}

// NewPlayer creates a new Player GameObject
func NewPlayer(position types.Vector2, size types.Vector2, moveSpeed float64) *Player {
	// Create the sprite (just handles texture and animation)
	playerSprite := sprite.NewSpriteSheet(
		config.Global.Player.TexturePath,
		sprite.Vector2{X: size.X, Y: size.Y},
		config.Global.Player.SpriteColumns,
		config.Global.Player.SpriteRows,
	)

	// Set animation speed from config
	playerSprite.SetFrameTime(config.Global.Animation.PlayerFrameTime)

	// Create the mover (handles position and velocity)
	playerMover := mover.NewBasicMover(
		position,
		types.Vector2{X: 0, Y: 0}, // Start stationary
		size.X,                    // Sprite width for wrapping
		size.Y,                    // Sprite height for wrapping
	)

	// Set screen bounds for wrapping from config
	playerMover.SetScreenBounds(config.Global.Screen.Width, config.Global.Screen.Height)

	return &Player{
		sprite:               playerSprite,
		mover:                playerMover,
		moveSpeed:            moveSpeed,
		debugMessageTimer:    0,
		debugMessageInterval: 2.0, // Post every 2 seconds
		actionTimer:          types.NewActionTimer(),
		stats: &types.EntityStats{
			HP:    100, // Will be overridden by config
			MaxHP: 100,
			Speed: 1.0,
		},
		selectedAction: types.ActionAttack, // Default action
		state: types.ObjectState{
			ID:       "Player",
			Position: position,
			Visible:  true,
			Frame:    0,
		},
	}
}

// GetSprite returns the sprite associated with this Player
func (p *Player) GetSprite() types.Sprite {
	return p.sprite
}

// GetMover returns the mover component for this Player
func (p *Player) GetMover() types.Mover {
	return p.mover
}

// GetState returns the Player's current state
func (p *Player) GetState() *types.ObjectState {
	return &p.state
}

// SetState sets the Player's state
func (p *Player) SetState(state types.ObjectState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = types.CopyObjectState(state)
}

// Update updates the Player's state
func (p *Player) Update(deltaTime float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update debug message timer
	p.debugMessageTimer += deltaTime

	// Post debug message periodically
	if p.debugMessageTimer >= p.debugMessageInterval {
		p.debugMessageTimer = 0
		pos := p.mover.GetPosition()
		types.PostDebugMessageSimple("Player", "Position: (%.0f, %.0f)", pos.X, pos.Y)
	}
}

// HandleInput updates the player's velocity based on input state
func (p *Player) HandleInput(inputState types.InputState) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate velocity based on input
	var velocityX, velocityY float64

	if inputState.MoveLeft {
		velocityX -= p.moveSpeed
	}
	if inputState.MoveRight {
		velocityX += p.moveSpeed
	}
	if inputState.MoveUp {
		velocityY -= p.moveSpeed
	}
	if inputState.MoveDown {
		velocityY += p.moveSpeed
	}

	// Normalize diagonal movement to maintain consistent speed
	if velocityX != 0 && velocityY != 0 {
		// Diagonal movement - reduce by sqrt(2) to maintain speed
		velocityX *= 0.7071 // 1/sqrt(2)
		velocityY *= 0.7071
	}

	// Update the mover's velocity
	p.mover.SetVelocity(types.Vector2{X: velocityX, Y: velocityY})
}

// GetID returns the player's unique identifier
func (p *Player) GetID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state.ID
}

// BattleEntity interface implementation

// GetActionTimer returns the player's action timer
func (p *Player) GetActionTimer() *types.ActionTimer {
	return p.actionTimer
}

// ChargeTimer charges the action timer by deltaTime
func (p *Player) ChargeTimer(deltaTime float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.actionTimer.Charge(deltaTime)
}

// ResetTimer resets the action timer to 0
func (p *Player) ResetTimer() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.actionTimer.Reset()
}

// IsReady returns true if the player can take an action
func (p *Player) IsReady() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.actionTimer.IsFull()
}

// GetStats returns the player's battle stats
func (p *Player) GetStats() *types.EntityStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stats
}

// SelectAction returns the player's selected action (from menu)
func (p *Player) SelectAction() *types.Action {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Player actions are selected through the menu system
	// This will be called when the timer is ready
	// For now, return nil - the menu system will handle action creation
	return nil
}

// SetSelectedAction sets the player's selected action from the menu
func (p *Player) SetSelectedAction(actionType types.ActionType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selectedAction = actionType
}

// GetSelectedAction returns the player's currently selected action
func (p *Player) GetSelectedAction() types.ActionType {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.selectedAction
}

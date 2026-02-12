//go:build js

package gameobject

import (
	"sync"

	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// BaseGameObject provides default implementations for the GameObject interface.
// Embed this struct in your custom GameObjects to avoid boilerplate code.
//
// Example usage:
//
//	type MyGameObject struct {
//	    *gameobject.BaseGameObject
//	    customField int
//	}
//
//	func NewMyGameObject(pos types.Vector2) *MyGameObject {
//	    return &MyGameObject{
//	        BaseGameObject: gameobject.NewBaseGameObject(sprite, mover, state),
//	        customField: 42,
//	    }
//	}
type BaseGameObject struct {
	sprite types.Sprite
	mover  types.Mover
	state  types.ObjectState
	mu     sync.Mutex
}

// NewBaseGameObject creates a new BaseGameObject with the given components.
// This is the recommended way to initialize a BaseGameObject.
func NewBaseGameObject(sprite types.Sprite, mover types.Mover, state types.ObjectState) *BaseGameObject {
	return &BaseGameObject{
		sprite: sprite,
		mover:  mover,
		state:  state,
	}
}

// GetSprite returns the sprite associated with this GameObject.
// Implements types.GameObject interface.
func (b *BaseGameObject) GetSprite() types.Sprite {
	return b.sprite
}

// GetMover returns the mover component, or nil if this object doesn't move.
// Implements types.GameObject interface.
func (b *BaseGameObject) GetMover() types.Mover {
	return b.mover
}

// GetState returns the GameObject's current state.
// Implements types.GameObject interface.
func (b *BaseGameObject) GetState() *types.ObjectState {
	return &b.state
}

// SetState sets the GameObject's state.
// Implements types.GameObject interface.
// This is thread-safe.
func (b *BaseGameObject) SetState(state types.ObjectState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = types.CopyObjectState(state)
}

// GetID returns the GameObject's unique identifier.
// Implements types.GameObject interface.
// This is thread-safe.
func (b *BaseGameObject) GetID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state.ID
}

// Update is a default no-op implementation.
// Override this in your custom GameObject if you need custom update logic.
// Implements types.GameObject interface.
func (b *BaseGameObject) Update(deltaTime float64) {
	// Default: no-op
	// Subclasses can override this method to add custom behavior
}

// SetSprite allows changing the sprite after construction.
// Useful for sprite swapping or dynamic sprite changes.
func (b *BaseGameObject) SetSprite(sprite types.Sprite) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sprite = sprite
}

// SetMover allows changing the mover after construction.
// Useful for changing movement behavior at runtime.
func (b *BaseGameObject) SetMover(mover types.Mover) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mover = mover
}


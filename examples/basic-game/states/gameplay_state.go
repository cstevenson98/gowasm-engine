package states

import (
	"example.com/basic-game/game/entities"
	"example.com/basic-game/game/gamestate"
	"github.com/cstevenson98/gowasm-engine/pkg/components"
	"github.com/cstevenson98/gowasm-engine/pkg/config"
	"github.com/cstevenson98/gowasm-engine/pkg/ecs"
	"github.com/cstevenson98/gowasm-engine/pkg/logger"
	"github.com/cstevenson98/gowasm-engine/pkg/prefab"
	"github.com/cstevenson98/gowasm-engine/pkg/state"
	"github.com/cstevenson98/gowasm-engine/pkg/systems"
	"github.com/cstevenson98/gowasm-engine/pkg/types"
)

// GameplayState is the walk-around gameplay state: a background plus a
// player-controlled entity, driven by input/movement/animation systems.
type GameplayState struct {
	*state.BaseState

	player ecs.Entity
}

// NewGameplayState creates the gameplay state.
func NewGameplayState() *GameplayState {
	return &GameplayState{BaseState: state.NewBaseState("Gameplay")}
}

// Enter builds the world (background + player) and registers systems.
func (s *GameplayState) Enter(deps state.Deps) error {
	if err := s.BaseState.Enter(deps); err != nil {
		return err
	}

	prefab.NewBackground(
		s.World(),
		types.Vector2{X: 0, Y: 0},
		types.Vector2{X: s.ScreenWidth(), Y: s.ScreenHeight()},
		"assets/art/test-background.png",
	)

	pos, stats := s.resolvePlayer()
	s.player = entities.SpawnPlayer(
		s.World(),
		pos,
		types.Vector2{X: config.Global.Player.Size, Y: config.Global.Player.Size},
		config.Global.Player.Speed,
		stats,
	)
	logger.Logger.Debugf("Spawned player at (%.2f, %.2f) in %s", pos.X, pos.Y, s.Name())

	// Order: input -> movement -> animation -> camera (so it follows the
	// player's post-movement position with no one-frame lag).
	s.Schedule().
		Add(entities.NewPlayerInputSystem(s.World())).
		Add(systems.NewMovement(s.World())).
		Add(systems.NewAnimation(s.World())).
		Add(systems.NewCameraFollow(s.World()))

	return nil
}

func (s *GameplayState) manager() *gamestate.GameStateManager {
	if m, ok := s.GameStateProvider().(*gamestate.GameStateManager); ok {
		return m
	}
	return nil
}

// resolvePlayer returns the spawn position and stats from the global game state
// (set on new game / load), falling back to configured defaults.
func (s *GameplayState) resolvePlayer() (types.Vector2, entities.Stats) {
	stats := entities.Stats{
		Level: 1,
		HP:    config.Global.Battle.PlayerHP,
		MaxHP: config.Global.Battle.PlayerMaxHP,
	}
	if m := s.manager(); m != nil {
		if gs := m.GetState(); gs != nil {
			stats = entities.Stats{
				Level:      gs.PlayerStats.Level,
				HP:         gs.PlayerStats.HP,
				MaxHP:      gs.PlayerStats.MaxHP,
				Experience: gs.PlayerStats.Experience,
			}
			return gs.PlayerPosition, stats
		}
	}
	spawnX, spawnY := config.GetPlayerSpawnPosition()
	return types.Vector2{X: spawnX, Y: spawnY}, stats
}

// Update runs systems then handles state switches.
func (s *GameplayState) Update(dt float64) {
	s.BaseState.Update(dt)

	in := s.Input()
	if in.Key2Pressed && !in.Key2PressedLastFrame {
		if err := s.RequestState(types.BATTLE); err != nil {
			logger.Logger.Errorf("Failed to switch to battle: %s", err.Error())
		}
	} else if in.MPressed && !in.MPressedLastFrame {
		if err := s.RequestState(types.PLAYER_MENU); err != nil {
			logger.Logger.Errorf("Failed to switch to player menu: %s", err.Error())
		}
	}
}

// Exit persists the player's live position back into the global game state so
// returning to gameplay (e.g. after the player menu) resumes in place.
func (s *GameplayState) Exit() {
	if m := s.manager(); m != nil {
		if p := ecs.NewMap1[components.Position](s.World()).Get(s.player); p != nil {
			m.UpdatePlayerPosition(types.Vector2{X: p.X, Y: p.Y})
		}
	}
	s.BaseState.Exit()
}

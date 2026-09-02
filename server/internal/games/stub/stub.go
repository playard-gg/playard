// Package stub provides a Game implementation that carries real metadata but
// has no rules yet. It lets the catalog, room, and lobby be built and tested
// generically before any game logic exists. Each entry here is deleted as the
// real game lands in its own package.
package stub

import (
	"errors"

	"github.com/kaviraj-j/playard/server/internal/games"
)

// ErrNotImplemented is returned by any attempt to actually play a stub game.
var ErrNotImplemented = errors.New("stub: this game has no rules yet")

type stubGame struct {
	meta games.Metadata
}

// New wraps metadata in a playable-shaped Game that refuses to play.
func New(meta games.Metadata) games.Game {
	return stubGame{meta: meta}
}

func (g stubGame) Metadata() games.Metadata { return g.meta }

func (g stubGame) Init(players []games.Player, _ games.Config) games.State {
	return map[string]any{"game_id": g.meta.ID, "players": players}
}

func (g stubGame) ApplyAction(state games.State, _ string, _ games.Action) (games.State, error) {
	return state, ErrNotImplemented
}

func (g stubGame) CheckEnd(games.State) *games.Result { return nil }

func (g stubGame) ViewFor(state games.State, _ string) games.ClientState { return state }

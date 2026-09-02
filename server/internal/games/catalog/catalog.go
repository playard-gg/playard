// Package catalog is the one place that lists the games this build ships.
// It lives outside the games package so games can stay dependency-free and
// individual game packages can import it without a cycle.
package catalog

import (
	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/stub"
)

// All returns every shipped game. Adding a game means adding a line here —
// the room, hub, and lobby need no changes.
func All() []games.Game {
	return []games.Game{
		stub.New(games.Metadata{
			ID:          "tic-tac-toe",
			Name:        "Tic Tac Toe",
			Tagline:     "Three in a row. Classic, quick, brutal.",
			Emoji:       "❌",
			MinPlayers:  2,
			MaxPlayers:  2,
			Pacing:      games.TurnBased,
			TeamMode:    games.TeamFFA,
			OutcomeType: games.OutcomeSingleWinner,
		}),
		stub.New(games.Metadata{
			ID:          "word-imposter",
			Name:        "Word Imposter",
			Tagline:     "Everyone gets the word. Almost everyone.",
			Emoji:       "🕵️",
			MinPlayers:  4,
			MaxPlayers:  8,
			Pacing:      games.TurnBased,
			TeamMode:    games.TeamFFA,
			OutcomeType: games.OutcomeSingleWinner,
		}),
		stub.New(games.Metadata{
			ID:          "quick-draw",
			Name:        "Quick Draw",
			Tagline:     "Fastest finger takes the round.",
			Emoji:       "⚡",
			MinPlayers:  2,
			MaxPlayers:  6,
			Pacing:      games.Realtime,
			TeamMode:    games.TeamFFA,
			OutcomeType: games.OutcomeRanked,
		}),
	}
}

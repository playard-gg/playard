// Package games defines the plug-and-play boundary between the room/hub
// layer and individual game implementations. The room never knows which
// game it is running — only that it satisfies the Game interface.
package games

import "encoding/json"

// Pacing describes how a game advances.
type Pacing string

const (
	TurnBased Pacing = "turn_based"
	Realtime  Pacing = "realtime"
)

// TeamMode describes how players are grouped.
type TeamMode string

const (
	TeamNone  TeamMode = "none"
	TeamFFA   TeamMode = "ffa"
	TeamFixed TeamMode = "fixed_teams"
)

// OutcomeType describes what "finished" means for a game.
type OutcomeType string

const (
	OutcomeNone         OutcomeType = "none"
	OutcomeSingleWinner OutcomeType = "single_winner"
	OutcomeRanked       OutcomeType = "ranked"
	OutcomeTeamWin      OutcomeType = "team_win"
)

// Metadata is everything the room/lobby needs to treat a game generically:
// validate player counts, render lobby chrome, and describe the game in the
// catalog. No game-specific logic leaks out beyond this.
type Metadata struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Tagline     string      `json:"tagline"`
	Emoji       string      `json:"emoji"`
	MinPlayers  int         `json:"min_players"`
	MaxPlayers  int         `json:"max_players"`
	Pacing      Pacing      `json:"pacing"`
	TeamMode    TeamMode    `json:"team_mode"`
	OutcomeType OutcomeType `json:"outcome_type"`
}

// Player is the game-facing view of a room member.
type Player struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

// Config carries per-room game options chosen in the lobby. Empty for now.
type Config map[string]any

// State is a game's internal, authoritative state. Only the owning game
// implementation interprets it.
type State any

// ClientState is the per-player view of State, safe to send over the wire.
type ClientState any

// Action is a raw player intent, decoded by the game implementation.
type Action json.RawMessage

// Result is the game-agnostic outcome the room reports once a game ends.
type Result struct {
	WinnerIDs []string       `json:"winner_ids,omitempty"`
	Ranking   []string       `json:"ranking,omitempty"`
	Points    map[string]int `json:"points,omitempty"`
}

// Game is the contract every game implements. The room/hub layer depends on
// this and nothing else.
type Game interface {
	Metadata() Metadata
	Init(players []Player, config Config) State
	ApplyAction(state State, playerID string, action Action) (State, error)
	CheckEnd(state State) *Result
	// ViewFor returns the view a specific player is allowed to see. Most
	// games return the same value for everyone; asymmetric-role games use
	// it to hide state without special-casing the room layer.
	ViewFor(state State, playerID string) ClientState
}

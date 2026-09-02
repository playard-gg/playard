package room

import (
	"errors"
	"sync"
	"time"

	"github.com/kaviraj-j/playard/server/internal/games"
)

// Status is where a room is in its lifecycle.
type Status string

const (
	StatusLobby    Status = "lobby"
	StatusInGame   Status = "in_game"
	StatusFinished Status = "finished"
)

// Visibility controls whether quickmatch can drop strangers into a room.
type Visibility string

const (
	Private Visibility = "private"
	Public  Visibility = "public"
)

var (
	ErrRoomFull       = errors.New("room: room is full")
	ErrGameInProgress = errors.New("room: game already in progress")
	ErrNotHost        = errors.New("room: only the host can do that")
	ErrNotEnough      = errors.New("room: not enough players")
	ErrNotInRoom      = errors.New("room: player is not in this room")
)

// Room is a lobby plus, once started, a running game instance. It owns its
// own lock; every exported method is safe for concurrent use.
type Room struct {
	mu sync.RWMutex

	code       string
	visibility Visibility
	game       games.Game
	meta       games.Metadata

	status    Status
	hostID    string
	players   []*Player
	gameState games.State
	result    *games.Result

	createdAt time.Time
	updatedAt time.Time
}

func newRoom(code string, game games.Game, visibility Visibility) *Room {
	now := time.Now()
	return &Room{
		code:       code,
		visibility: visibility,
		game:       game,
		meta:       game.Metadata(),
		status:     StatusLobby,
		createdAt:  now,
		updatedAt:  now,
	}
}

// Code returns the room's join code.
func (r *Room) Code() string { return r.code }

// GameID returns the id of the game this room is for.
func (r *Room) GameID() string { return r.meta.ID }

// Join adds a player, or reclaims an existing seat if the same player id is
// already present (reconnect, or a second tab).
func (r *Room) Join(playerID, nickname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p := r.findLocked(playerID); p != nil {
		p.Nickname = nickname
		r.touchLocked()
		return nil
	}
	if r.status != StatusLobby {
		return ErrGameInProgress
	}
	if len(r.players) >= r.meta.MaxPlayers {
		return ErrRoomFull
	}

	// DisconnectedAt starts at the join time so a player who joins but never
	// completes their socket handshake is reaped after the normal grace
	// period, rather than being eligible for reaping immediately.
	r.players = append(r.players, &Player{ID: playerID, Nickname: nickname, DisconnectedAt: time.Now()})
	if r.hostID == "" {
		r.hostID = playerID
	}
	r.touchLocked()
	return nil
}

// Leave removes a player outright (an explicit leave, not a dropped socket).
// If the host leaves, the next remaining player is promoted.
func (r *Room) Leave(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.players {
		if p.ID != playerID {
			continue
		}
		r.players = append(r.players[:i], r.players[i+1:]...)
		break
	}
	if r.hostID == playerID && len(r.players) > 0 {
		r.hostID = r.players[0].ID
	}
	r.touchLocked()
}

// SetConnected marks a player's socket as present or gone. A gone socket does
// not free the seat — the manager's reaper does that after the grace period.
func (r *Room) SetConnected(playerID string, connected bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := r.findLocked(playerID)
	if p == nil {
		return
	}
	p.Connected = connected
	if connected {
		p.DisconnectedAt = time.Time{}
	} else {
		p.DisconnectedAt = time.Now()
	}
	r.touchLocked()
}

// Start transitions the room into its game. Only the host may start, and only
// once enough players are present — being in the room is being ready.
func (r *Room) Start(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.status != StatusLobby {
		return ErrGameInProgress
	}
	if playerID != r.hostID {
		return ErrNotHost
	}
	if len(r.players) < r.meta.MinPlayers {
		return ErrNotEnough
	}

	gamePlayers := make([]games.Player, 0, len(r.players))
	for _, p := range r.players {
		gamePlayers = append(gamePlayers, games.Player{ID: p.ID, Nickname: p.Nickname})
	}

	r.gameState = r.game.Init(gamePlayers, games.Config{})
	r.status = StatusInGame
	r.touchLocked()
	return nil
}

func (r *Room) findLocked(playerID string) *Player {
	for _, p := range r.players {
		if p.ID == playerID {
			return p
		}
	}
	return nil
}

func (r *Room) touchLocked() { r.updatedAt = time.Now() }

// View is the wire snapshot of a room, sent to clients on every change.
type View struct {
	Code       string            `json:"code"`
	GameID     string            `json:"game_id"`
	GameName   string            `json:"game_name"`
	Visibility Visibility        `json:"visibility"`
	Status     Status            `json:"status"`
	HostID     string            `json:"host_id"`
	MinPlayers int               `json:"min_players"`
	MaxPlayers int               `json:"max_players"`
	Players    []PlayerView      `json:"players"`
	CanStart   bool              `json:"can_start"`
	Result     *games.Result     `json:"result,omitempty"`
	GameView   games.ClientState `json:"game_view,omitempty"`
}

// ViewFor renders the room from one player's perspective. The game-specific
// portion is delegated to the game so asymmetric-role games can hide state
// without the room layer knowing anything about them.
func (r *Room) ViewFor(playerID string) View {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]PlayerView, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, PlayerView{
			ID:        p.ID,
			Nickname:  p.Nickname,
			Connected: p.Connected,
			IsHost:    p.ID == r.hostID,
		})
	}

	view := View{
		Code:       r.code,
		GameID:     r.meta.ID,
		GameName:   r.meta.Name,
		Visibility: r.visibility,
		Status:     r.status,
		HostID:     r.hostID,
		MinPlayers: r.meta.MinPlayers,
		MaxPlayers: r.meta.MaxPlayers,
		Players:    players,
		CanStart:   r.canStartLocked(),
		Result:     r.result,
	}
	if r.status == StatusInGame && r.gameState != nil {
		view.GameView = r.game.ViewFor(r.gameState, playerID)
	}
	return view
}

// Summary is the pre-join preview returned over HTTP, before a player has a
// socket. It deliberately exposes nothing about game state.
type Summary struct {
	Code        string `json:"code"`
	GameID      string `json:"game_id"`
	GameName    string `json:"game_name"`
	Status      Status `json:"status"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
}

// Summarize returns the pre-join preview for this room.
func (r *Room) Summarize() Summary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return Summary{
		Code:        r.code,
		GameID:      r.meta.ID,
		GameName:    r.meta.Name,
		Status:      r.status,
		PlayerCount: len(r.players),
		MaxPlayers:  r.meta.MaxPlayers,
	}
}

func (r *Room) canStartLocked() bool {
	return r.status == StatusLobby && len(r.players) >= r.meta.MinPlayers
}

// openForQuickmatch reports whether strangers may be dropped into this room.
func (r *Room) openForQuickmatch() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.visibility == Public &&
		r.status == StatusLobby &&
		len(r.players) < r.meta.MaxPlayers
}

// expired reports whether a room can be reaped: empty, or holding only
// disconnected seats, for longer than the grace period.
func (r *Room) expired(grace time.Duration, now time.Time) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.players) == 0 {
		return now.Sub(r.updatedAt) > grace
	}
	for _, p := range r.players {
		if p.Connected {
			return false
		}
		if now.Sub(p.DisconnectedAt) <= grace {
			return false
		}
	}
	return true
}

// dropExpiredSeats removes players whose grace period has run out, returning
// true if anything changed so the caller can rebroadcast.
func (r *Room) dropExpiredSeats(grace time.Duration, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	kept := r.players[:0]
	changed := false
	for _, p := range r.players {
		if !p.Connected && now.Sub(p.DisconnectedAt) > grace {
			changed = true
			continue
		}
		kept = append(kept, p)
	}
	r.players = kept

	if changed {
		if r.findLocked(r.hostID) == nil && len(r.players) > 0 {
			r.hostID = r.players[0].ID
		}
		r.touchLocked()
	}
	return changed
}

// playersSnapshot returns a copy of the player slice for read-only use by the
// manager, which must not hold a room's lock across its own map lock.
func (r *Room) playersSnapshot() []*Player {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Player, len(r.players))
	copy(out, r.players)
	return out
}

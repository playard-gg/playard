package room

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kaviraj-j/playard/server/internal/games"
)

// ErrRoomNotFound is returned for an unknown or already-reaped room code.
var ErrRoomNotFound = errors.New("room: not found")

const (
	// DefaultGrace is how long a disconnected player's seat is held before it
	// is freed, and how long an empty room lingers before being reaped.
	DefaultGrace = 45 * time.Second
	// reapInterval is how often expired rooms and seats are swept.
	reapInterval = 15 * time.Second
	// codeAttempts bounds retries when a generated code already exists.
	codeAttempts = 10
)

// Manager owns every live room. Rooms exist only in memory — nothing about a
// room outlives the process, which matches the no-accounts model.
type Manager struct {
	mu       sync.RWMutex
	rooms    map[string]*Room
	registry *games.Registry
	grace    time.Duration

	// onChange is called after the reaper mutates a room, so the ws layer can
	// rebroadcast without the room package importing it.
	onChange func(*Room)
}

// NewManager constructs a room manager over the given game registry.
func NewManager(registry *games.Registry, grace time.Duration) *Manager {
	if grace <= 0 {
		grace = DefaultGrace
	}
	return &Manager{
		rooms:    make(map[string]*Room),
		registry: registry,
		grace:    grace,
	}
}

// OnChange registers a callback invoked whenever the reaper changes a room.
func (m *Manager) OnChange(fn func(*Room)) { m.onChange = fn }

// Grace is how long a disconnected player keeps their seat.
func (m *Manager) Grace() time.Duration { return m.grace }

// Create makes a new room for the given game.
func (m *Manager) Create(gameID string, visibility Visibility) (*Room, error) {
	game, err := m.registry.Get(gameID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < codeAttempts; i++ {
		code, err := NewCode()
		if err != nil {
			return nil, err
		}
		if _, taken := m.rooms[code]; taken {
			continue
		}
		r := newRoom(code, game, visibility)
		m.rooms[code] = r
		return r, nil
	}
	return nil, fmt.Errorf("room: could not allocate a free code after %d attempts", codeAttempts)
}

// Get looks up a room by code, normalizing user-typed input first.
func (m *Manager) Get(code string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.rooms[NormalizeCode(code)]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

// Quickmatch finds a public room for the game that is still in its lobby and
// has space, creating one if none exists. This is deliberately not a queue:
// the player is in a room immediately either way.
func (m *Manager) Quickmatch(gameID string) (*Room, error) {
	if _, err := m.registry.Get(gameID); err != nil {
		return nil, err
	}

	m.mu.RLock()
	var best *Room
	for _, r := range m.rooms {
		if r.GameID() != gameID || !r.openForQuickmatch() {
			continue
		}
		// Prefer the fullest open room so games fill up and start sooner
		// instead of scattering players across half-empty lobbies.
		if best == nil || len(r.playersSnapshot()) > len(best.playersSnapshot()) {
			best = r
		}
	}
	m.mu.RUnlock()

	if best != nil {
		return best, nil
	}
	return m.Create(gameID, Public)
}

// StartReaper sweeps expired seats and rooms until done is closed.
func (m *Manager) StartReaper(done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(reapInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case now := <-ticker.C:
				m.reap(now)
			}
		}
	}()
}

func (m *Manager) reap(now time.Time) {
	m.mu.RLock()
	snapshot := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		snapshot = append(snapshot, r)
	}
	m.mu.RUnlock()

	var dead []string
	for _, r := range snapshot {
		if r.expired(m.grace, now) {
			dead = append(dead, r.Code())
			continue
		}
		if r.dropExpiredSeats(m.grace, now) && m.onChange != nil {
			m.onChange(r)
		}
	}
	if len(dead) == 0 {
		return
	}

	m.mu.Lock()
	for _, code := range dead {
		delete(m.rooms, code)
	}
	m.mu.Unlock()
}

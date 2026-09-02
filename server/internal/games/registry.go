package games

import (
	"errors"
	"fmt"
	"sort"
)

// ErrUnknownGame is returned when a game id has no registered implementation.
var ErrUnknownGame = errors.New("games: unknown game")

// Registry holds the available games. It is built once at startup and read
// concurrently thereafter, so it needs no lock.
type Registry struct {
	byID map[string]Game
}

// NewRegistry builds a registry from the given games, erroring on duplicate
// or invalid metadata rather than silently shadowing an entry.
func NewRegistry(list ...Game) (*Registry, error) {
	byID := make(map[string]Game, len(list))
	for _, g := range list {
		meta := g.Metadata()
		if meta.ID == "" {
			return nil, errors.New("games: game registered with empty id")
		}
		if meta.MinPlayers < 1 || meta.MaxPlayers < meta.MinPlayers {
			return nil, fmt.Errorf("games: %s has invalid player range %d-%d", meta.ID, meta.MinPlayers, meta.MaxPlayers)
		}
		if _, exists := byID[meta.ID]; exists {
			return nil, fmt.Errorf("games: duplicate game id %q", meta.ID)
		}
		byID[meta.ID] = g
	}
	return &Registry{byID: byID}, nil
}

// Get returns the game registered under id.
func (r *Registry) Get(id string) (Game, error) {
	g, ok := r.byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownGame, id)
	}
	return g, nil
}

// Metadata returns the metadata for a single game.
func (r *Registry) Metadata(id string) (Metadata, error) {
	g, err := r.Get(id)
	if err != nil {
		return Metadata{}, err
	}
	return g.Metadata(), nil
}

// All returns every game's metadata, sorted by id for a stable catalog.
func (r *Registry) All() []Metadata {
	out := make([]Metadata, 0, len(r.byID))
	for _, g := range r.byID {
		out = append(out, g.Metadata())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

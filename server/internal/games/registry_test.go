package games_test

import (
	"errors"
	"testing"

	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/catalog"
	"github.com/kaviraj-j/playard/server/internal/games/stub"
)

func meta(id string, min, max int) games.Metadata {
	return games.Metadata{ID: id, Name: id, MinPlayers: min, MaxPlayers: max}
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name    string
		list    []games.Game
		wantErr bool
	}{
		{"empty registry is valid", nil, false},
		{"single game", []games.Game{stub.New(meta("a", 2, 4))}, false},
		{"duplicate id", []games.Game{stub.New(meta("a", 2, 4)), stub.New(meta("a", 2, 4))}, true},
		{"empty id", []games.Game{stub.New(meta("", 2, 4))}, true},
		{"max below min", []games.Game{stub.New(meta("a", 4, 2))}, true},
		{"zero min", []games.Game{stub.New(meta("a", 0, 4))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := games.NewRegistry(tt.list...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewRegistry() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryGet(t *testing.T) {
	r, err := games.NewRegistry(stub.New(meta("a", 2, 4)))
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	if _, err := r.Get("a"); err != nil {
		t.Errorf("Get(a) err = %v, want nil", err)
	}
	if _, err := r.Get("missing"); !errors.Is(err, games.ErrUnknownGame) {
		t.Errorf("Get(missing) err = %v, want ErrUnknownGame", err)
	}
}

func TestRegistryAllIsSorted(t *testing.T) {
	r, _ := games.NewRegistry(stub.New(meta("c", 2, 4)), stub.New(meta("a", 2, 4)), stub.New(meta("b", 2, 4)))

	all := r.All()
	want := []string{"a", "b", "c"}
	if len(all) != len(want) {
		t.Fatalf("All() returned %d games, want %d", len(all), len(want))
	}
	for i, m := range all {
		if m.ID != want[i] {
			t.Errorf("All()[%d].ID = %q, want %q", i, m.ID, want[i])
		}
	}
}

// The shipped catalog must be valid, or the server refuses to start.
func TestShippedCatalogIsValid(t *testing.T) {
	r, err := games.NewRegistry(catalog.All()...)
	if err != nil {
		t.Fatalf("shipped catalog is invalid: %v", err)
	}
	if len(r.All()) == 0 {
		t.Fatal("shipped catalog is empty")
	}
	for _, m := range r.All() {
		if m.Name == "" || m.Emoji == "" || m.Tagline == "" {
			t.Errorf("game %q is missing catalog display fields", m.ID)
		}
	}
}

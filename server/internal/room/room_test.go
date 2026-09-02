package room

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/stub"
)

func testGame(min, max int) games.Game {
	return stub.New(games.Metadata{
		ID:          "test-game",
		Name:        "Test Game",
		MinPlayers:  min,
		MaxPlayers:  max,
		Pacing:      games.TurnBased,
		TeamMode:    games.TeamFFA,
		OutcomeType: games.OutcomeSingleWinner,
	})
}

func testManager(t *testing.T, min, max int) *Manager {
	t.Helper()
	registry, err := games.NewRegistry(testGame(min, max))
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	return NewManager(registry, DefaultGrace)
}

func TestNewCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 500; i++ {
		code, err := NewCode()
		if err != nil {
			t.Fatalf("NewCode: %v", err)
		}
		if len(code) != CodeLength {
			t.Fatalf("code %q has length %d, want %d", code, len(code), CodeLength)
		}
		for _, c := range code {
			if !strings.ContainsRune(codeAlphabet, c) {
				t.Fatalf("code %q contains %q which is outside the alphabet", code, c)
			}
		}
		if seen[code] {
			t.Fatalf("duplicate code %q within 500 draws", code)
		}
		seen[code] = true
	}
}

func TestNormalizeCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already normal", "ABC234", "ABC234"},
		{"lowercase", "abc234", "ABC234"},
		{"padded", "  abc234\n", "ABC234"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCode(tt.input); got != tt.want {
				t.Errorf("NormalizeCode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	t.Run("first player becomes host", func(t *testing.T) {
		r := newRoom("ABC234", testGame(2, 4), Private)
		if err := r.Join("p1", "One"); err != nil {
			t.Fatalf("join: %v", err)
		}
		if err := r.Join("p2", "Two"); err != nil {
			t.Fatalf("join: %v", err)
		}
		if got := r.ViewFor("p1").HostID; got != "p1" {
			t.Errorf("host = %q, want p1", got)
		}
	})

	t.Run("rejoining the same id reclaims the seat", func(t *testing.T) {
		r := newRoom("ABC234", testGame(2, 4), Private)
		r.Join("p1", "One")
		if err := r.Join("p1", "One Renamed"); err != nil {
			t.Fatalf("rejoin: %v", err)
		}
		view := r.ViewFor("p1")
		if len(view.Players) != 1 {
			t.Fatalf("players = %d, want 1", len(view.Players))
		}
		if view.Players[0].Nickname != "One Renamed" {
			t.Errorf("nickname = %q, want the updated one", view.Players[0].Nickname)
		}
	})

	t.Run("full room is rejected", func(t *testing.T) {
		r := newRoom("ABC234", testGame(2, 2), Private)
		r.Join("p1", "One")
		r.Join("p2", "Two")
		if err := r.Join("p3", "Three"); !errors.Is(err, ErrRoomFull) {
			t.Errorf("err = %v, want ErrRoomFull", err)
		}
	})

	t.Run("in-progress room is rejected", func(t *testing.T) {
		r := newRoom("ABC234", testGame(2, 4), Private)
		r.Join("p1", "One")
		r.Join("p2", "Two")
		if err := r.Start("p1"); err != nil {
			t.Fatalf("start: %v", err)
		}
		if err := r.Join("p3", "Three"); !errors.Is(err, ErrGameInProgress) {
			t.Errorf("err = %v, want ErrGameInProgress", err)
		}
	})
}

func TestStart(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Room)
		starter string
		wantErr error
	}{
		{
			name: "host starts a full-enough room",
			setup: func(r *Room) {
				r.Join("p1", "One")
				r.Join("p2", "Two")
			},
			starter: "p1",
			wantErr: nil,
		},
		{
			name: "non-host cannot start",
			setup: func(r *Room) {
				r.Join("p1", "One")
				r.Join("p2", "Two")
			},
			starter: "p2",
			wantErr: ErrNotHost,
		},
		{
			name: "below minimum players",
			setup: func(r *Room) {
				r.Join("p1", "One")
			},
			starter: "p1",
			wantErr: ErrNotEnough,
		},
		{
			name: "already started",
			setup: func(r *Room) {
				r.Join("p1", "One")
				r.Join("p2", "Two")
				r.Start("p1")
			},
			starter: "p1",
			wantErr: ErrGameInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRoom("ABC234", testGame(2, 4), Private)
			tt.setup(r)

			err := r.Start(tt.starter)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Start() err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got := r.ViewFor("p1").Status; got != StatusInGame {
				t.Errorf("status = %q, want in_game", got)
			}
		})
	}
}

func TestLeavePromotesNewHost(t *testing.T) {
	r := newRoom("ABC234", testGame(2, 4), Private)
	r.Join("p1", "One")
	r.Join("p2", "Two")

	r.Leave("p1")

	view := r.ViewFor("p2")
	if view.HostID != "p2" {
		t.Errorf("host = %q, want p2", view.HostID)
	}
	if len(view.Players) != 1 {
		t.Errorf("players = %d, want 1", len(view.Players))
	}
}

func TestDisconnectHoldsSeatUntilGrace(t *testing.T) {
	grace := 50 * time.Millisecond
	r := newRoom("ABC234", testGame(2, 4), Private)
	r.Join("p1", "One")
	r.Join("p2", "Two")
	r.SetConnected("p1", true)
	r.SetConnected("p2", true)

	r.SetConnected("p2", false)

	if r.dropExpiredSeats(grace, time.Now()) {
		t.Fatal("seat was freed immediately; it should be held for the grace period")
	}
	if got := len(r.ViewFor("p1").Players); got != 2 {
		t.Fatalf("players = %d, want 2 while the seat is held", got)
	}

	if !r.dropExpiredSeats(grace, time.Now().Add(2*grace)) {
		t.Fatal("seat was not freed after the grace period")
	}
	if got := len(r.ViewFor("p1").Players); got != 1 {
		t.Errorf("players = %d, want 1 after the grace period", got)
	}
}

func TestQuickmatch(t *testing.T) {
	t.Run("creates a public room when none is waiting", func(t *testing.T) {
		m := testManager(t, 2, 4)
		r, err := m.Quickmatch("test-game")
		if err != nil {
			t.Fatalf("quickmatch: %v", err)
		}
		if r.ViewFor("").Visibility != Public {
			t.Error("quickmatch room should be public")
		}
	})

	t.Run("reuses an open public room", func(t *testing.T) {
		m := testManager(t, 2, 4)
		first, _ := m.Quickmatch("test-game")
		first.Join("p1", "One")

		second, err := m.Quickmatch("test-game")
		if err != nil {
			t.Fatalf("quickmatch: %v", err)
		}
		if second.Code() != first.Code() {
			t.Errorf("got a new room %q, want the waiting room %q", second.Code(), first.Code())
		}
	})

	t.Run("skips full rooms", func(t *testing.T) {
		m := testManager(t, 2, 2)
		first, _ := m.Quickmatch("test-game")
		first.Join("p1", "One")
		first.Join("p2", "Two")

		second, _ := m.Quickmatch("test-game")
		if second.Code() == first.Code() {
			t.Error("quickmatch put a player into a full room")
		}
	})

	t.Run("never joins a private room", func(t *testing.T) {
		m := testManager(t, 2, 4)
		private, _ := m.Create("test-game", Private)
		private.Join("p1", "One")

		matched, _ := m.Quickmatch("test-game")
		if matched.Code() == private.Code() {
			t.Error("quickmatch leaked a player into a private room")
		}
	})

	t.Run("unknown game", func(t *testing.T) {
		m := testManager(t, 2, 4)
		if _, err := m.Quickmatch("nope"); !errors.Is(err, games.ErrUnknownGame) {
			t.Errorf("err = %v, want ErrUnknownGame", err)
		}
	})
}

func TestManagerGetIsCaseInsensitive(t *testing.T) {
	m := testManager(t, 2, 4)
	created, _ := m.Create("test-game", Private)

	got, err := m.Get(strings.ToLower("  " + created.Code() + " "))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Code() != created.Code() {
		t.Errorf("code = %q, want %q", got.Code(), created.Code())
	}

	if _, err := m.Get("ZZZZZZ"); !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("err = %v, want ErrRoomNotFound", err)
	}
}

func TestReapRemovesEmptyRooms(t *testing.T) {
	registry, _ := games.NewRegistry(testGame(2, 4))
	m := NewManager(registry, 10*time.Millisecond)
	created, _ := m.Create("test-game", Private)

	m.reap(time.Now())
	if _, err := m.Get(created.Code()); err != nil {
		t.Fatal("room was reaped before its grace period elapsed")
	}

	m.reap(time.Now().Add(time.Second))
	if _, err := m.Get(created.Code()); !errors.Is(err, ErrRoomNotFound) {
		t.Errorf("err = %v, want the empty room to be reaped", err)
	}
}

package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/stub"
	"github.com/kaviraj-j/playard/server/internal/httpapi"
	"github.com/kaviraj-j/playard/server/internal/room"
)

type harness struct {
	mux   *http.ServeMux
	rooms *room.Manager
	token string
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	registry, err := games.NewRegistry(stub.New(games.Metadata{
		ID: "test-game", Name: "Test Game", MinPlayers: 2, MaxPlayers: 2,
	}))
	if err != nil {
		t.Fatalf("registry: %v", err)
	}

	authService := auth.NewService([]byte("test-secret"))
	token, err := authService.Issue("p1", "One")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	rooms := room.NewManager(registry, room.DefaultGrace)
	mux := http.NewServeMux()
	passthrough := func(h http.HandlerFunc) http.HandlerFunc { return h }
	httpapi.New(registry, rooms).Routes(mux, authService, passthrough)

	return &harness{mux: mux, rooms: rooms, token: token}
}

func (h *harness) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return out
}

func TestListGamesNeedsNoAuth(t *testing.T) {
	h := newHarness(t)
	rec := h.do(t, http.MethodGet, "/api/games", nil, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := decode[struct {
		Games []games.Metadata `json:"games"`
	}](t, rec)
	if len(body.Games) != 1 || body.Games[0].ID != "test-game" {
		t.Errorf("games = %+v, want the single test game", body.Games)
	}
}

func TestRoomEndpointsRequireAuth(t *testing.T) {
	h := newHarness(t)
	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"create", http.MethodPost, "/api/rooms", map[string]string{"game_id": "test-game"}},
		{"join", http.MethodPost, "/api/rooms/join", map[string]string{"code": "ABC234"}},
		{"quickmatch", http.MethodPost, "/api/rooms/quickmatch", map[string]string{"game_id": "test-game"}},
		{"preview", http.MethodGet, "/api/rooms/ABC234", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := h.do(t, tt.method, tt.path, tt.body, ""); rec.Code != http.StatusUnauthorized {
				t.Errorf("without a token status = %d, want 401", rec.Code)
			}
			if rec := h.do(t, tt.method, tt.path, tt.body, "garbage"); rec.Code != http.StatusUnauthorized {
				t.Errorf("with a bad token status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestCreateRoom(t *testing.T) {
	h := newHarness(t)

	rec := h.do(t, http.MethodPost, "/api/rooms", map[string]string{"game_id": "test-game", "visibility": "private"}, h.token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	body := decode[struct {
		Code   string `json:"code"`
		GameID string `json:"game_id"`
	}](t, rec)
	if len(body.Code) != room.CodeLength {
		t.Errorf("code = %q, want %d characters", body.Code, room.CodeLength)
	}
	if body.GameID != "test-game" {
		t.Errorf("game_id = %q, want test-game", body.GameID)
	}

	if rec := h.do(t, http.MethodPost, "/api/rooms", map[string]string{"game_id": "nope"}, h.token); rec.Code != http.StatusNotFound {
		t.Errorf("unknown game status = %d, want 404", rec.Code)
	}
}

func TestJoinRoom(t *testing.T) {
	h := newHarness(t)
	created, err := h.rooms.Create("test-game", room.Private)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	t.Run("valid code, lowercased", func(t *testing.T) {
		rec := h.do(t, http.MethodPost, "/api/rooms/join", map[string]string{"code": created.Code()}, h.token)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("unknown code", func(t *testing.T) {
		rec := h.do(t, http.MethodPost, "/api/rooms/join", map[string]string{"code": "ZZZZZZ"}, h.token)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("full room", func(t *testing.T) {
		created.Join("a", "A")
		created.Join("b", "B")
		rec := h.do(t, http.MethodPost, "/api/rooms/join", map[string]string{"code": created.Code()}, h.token)
		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("already started", func(t *testing.T) {
		started, _ := h.rooms.Create("test-game", room.Private)
		started.Join("a", "A")
		started.Join("b", "B")
		if err := started.Start("a"); err != nil {
			t.Fatalf("start: %v", err)
		}
		rec := h.do(t, http.MethodPost, "/api/rooms/join", map[string]string{"code": started.Code()}, h.token)
		if rec.Code != http.StatusConflict {
			t.Errorf("status = %d, want 409", rec.Code)
		}
	})
}

func TestQuickmatchReusesWaitingRoom(t *testing.T) {
	h := newHarness(t)

	first := decode[struct {
		Code string `json:"code"`
	}](t, h.do(t, http.MethodPost, "/api/rooms/quickmatch", map[string]string{"game_id": "test-game"}, h.token))

	second := decode[struct {
		Code string `json:"code"`
	}](t, h.do(t, http.MethodPost, "/api/rooms/quickmatch", map[string]string{"game_id": "test-game"}, h.token))

	if first.Code != second.Code {
		t.Errorf("quickmatch created a second room (%q, %q) instead of reusing the waiting one", first.Code, second.Code)
	}
}

func TestGetRoomPreview(t *testing.T) {
	h := newHarness(t)
	created, _ := h.rooms.Create("test-game", room.Private)
	created.Join("a", "A")

	rec := h.do(t, http.MethodGet, "/api/rooms/"+created.Code(), nil, h.token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	summary := decode[room.Summary](t, rec)
	if summary.PlayerCount != 1 || summary.GameName != "Test Game" || summary.Status != room.StatusLobby {
		t.Errorf("summary = %+v, want 1 player in the Test Game lobby", summary)
	}
}

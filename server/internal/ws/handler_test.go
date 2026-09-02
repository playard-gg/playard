package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/stub"
	"github.com/kaviraj-j/playard/server/internal/room"
	"github.com/kaviraj-j/playard/server/internal/ws"
)

type roomView struct {
	Code     string `json:"code"`
	Status   string `json:"status"`
	HostID   string `json:"host_id"`
	CanStart bool   `json:"can_start"`
	Players  []struct {
		ID        string `json:"id"`
		Nickname  string `json:"nickname"`
		Connected bool   `json:"connected"`
		IsHost    bool   `json:"is_host"`
	} `json:"players"`
}

type frame struct {
	Type string   `json:"type"`
	Data roomView `json:"data"`
}

type errFrame struct {
	Type string `json:"type"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

type fixture struct {
	server *httptest.Server
	auth   *auth.Service
	rooms  *room.Manager
}

func newFixture(t *testing.T, minPlayers, maxPlayers int) *fixture {
	t.Helper()

	registry, err := games.NewRegistry(stub.New(games.Metadata{
		ID: "test-game", Name: "Test Game", MinPlayers: minPlayers, MaxPlayers: maxPlayers,
	}))
	if err != nil {
		t.Fatalf("registry: %v", err)
	}

	authService := auth.NewService([]byte("test-secret"))
	rooms := room.NewManager(registry, room.DefaultGrace)
	hubs := ws.NewRegistry()
	rooms.OnChange(hubs.Broadcast)

	srv := httptest.NewServer(ws.NewHandler(authService, rooms, hubs, "*"))
	t.Cleanup(srv.Close)

	return &fixture{server: srv, auth: authService, rooms: rooms}
}

// dial connects a player to a room, failing the test if the handshake does not
// succeed.
func (f *fixture) dial(t *testing.T, code, playerID, nickname string) *gorilla.Conn {
	t.Helper()

	c, resp, err := f.tryDial(code, playerID, nickname)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("dial: %v (status %d)", err, status)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func (f *fixture) tryDial(code, playerID, nickname string) (*gorilla.Conn, *http.Response, error) {
	token, err := f.auth.Issue(playerID, nickname)
	if err != nil {
		return nil, nil, err
	}
	url := "ws" + strings.TrimPrefix(f.server.URL, "http") + "/api/ws?code=" + code + "&token=" + token
	return gorilla.DefaultDialer.Dial(url, nil)
}

// readRaw waits for the next message as raw JSON.
func readRaw(t *testing.T, c *gorilla.Conn) []byte {
	t.Helper()
	c.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	return data
}

// readFrame waits for the next message, decoded as a room state frame.
func readFrame(t *testing.T, c *gorilla.Conn) frame {
	t.Helper()

	var f frame
	if err := json.Unmarshal(readRaw(t, c), &f); err != nil {
		t.Fatalf("decode frame: %v", err)
	}
	return f
}

// awaitState reads until a room-state frame satisfies want. Broadcasts queued
// by other players connecting or leaving can arrive in any order relative to
// the action under test, so tests wait for the condition rather than assuming
// the very next frame carries it.
func awaitState(t *testing.T, c *gorilla.Conn, what string, want func(roomView) bool) frame {
	t.Helper()
	for i := 0; i < 20; i++ {
		f := readFrame(t, c)
		if (f.Type == ws.MsgRoomState || f.Type == ws.MsgGameStarted) && want(f.Data) {
			return f
		}
	}
	t.Fatalf("never observed room state where %s", what)
	return frame{}
}

// awaitError reads until an error frame arrives, returning its message.
func awaitError(t *testing.T, c *gorilla.Conn) string {
	t.Helper()
	for i := 0; i < 20; i++ {
		var e errFrame
		if err := json.Unmarshal(readRaw(t, c), &e); err != nil {
			t.Fatalf("decode frame: %v", err)
		}
		if e.Type == ws.MsgError {
			return e.Data.Message
		}
	}
	t.Fatal("never received an error frame")
	return ""
}

// readUntil waits for a frame of the given type, skipping intermediate
// broadcasts caused by other players connecting.
func readUntil(t *testing.T, c *gorilla.Conn, msgType string) frame {
	t.Helper()
	for i := 0; i < 20; i++ {
		if f := readFrame(t, c); f.Type == msgType {
			return f
		}
	}
	t.Fatalf("never received a %q frame", msgType)
	return frame{}
}

// playerByID reports whether a player is present in the view, and if so
// whether they are currently connected.
func playerByID(view roomView, id string) (present, connected bool) {
	for _, p := range view.Players {
		if p.ID == id {
			return true, p.Connected
		}
	}
	return false, false
}

func send(t *testing.T, c *gorilla.Conn, msgType string) {
	t.Helper()
	if err := c.WriteJSON(map[string]string{"type": msgType}); err != nil {
		t.Fatalf("send %s: %v", msgType, err)
	}
}

func TestConnectRejectsBadHandshake(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	t.Run("invalid token", func(t *testing.T) {
		url := "ws" + strings.TrimPrefix(f.server.URL, "http") + "/api/ws?code=" + rm.Code() + "&token=garbage"
		_, resp, err := gorilla.DefaultDialer.Dial(url, nil)
		if err == nil {
			t.Fatal("dial with a garbage token succeeded")
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", resp.StatusCode)
		}
	})

	t.Run("unknown room", func(t *testing.T) {
		_, resp, err := f.tryDial("ZZZZZZ", "p1", "One")
		if err == nil {
			t.Fatal("dial into a nonexistent room succeeded")
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
	})

	t.Run("full room", func(t *testing.T) {
		small := newFixture(t, 1, 1)
		full, _ := small.rooms.Create("test-game", room.Private)
		small.dial(t, full.Code(), "p1", "One")

		_, resp, err := small.tryDial(full.Code(), "p2", "Two")
		if err == nil {
			t.Fatal("dial into a full room succeeded")
		}
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("status = %d, want 403", resp.StatusCode)
		}
	})
}

func TestConnectBroadcastsRoomState(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	first := readFrame(t, host)
	if first.Type != ws.MsgRoomState {
		t.Fatalf("first frame type = %q, want room_state", first.Type)
	}
	if len(first.Data.Players) != 1 || first.Data.HostID != "p1" {
		t.Fatalf("first frame = %+v, want a single host player", first.Data)
	}
	if !first.Data.Players[0].Connected {
		t.Error("the connected player is not marked connected")
	}

	guest := f.dial(t, rm.Code(), "p2", "Two")

	// Both sockets see the same two-player room.
	hostView := readUntil(t, host, ws.MsgRoomState)
	guestView := readUntil(t, guest, ws.MsgRoomState)
	if len(hostView.Data.Players) != 2 {
		t.Errorf("host sees %d players, want 2", len(hostView.Data.Players))
	}
	if len(guestView.Data.Players) != 2 {
		t.Errorf("guest sees %d players, want 2", len(guestView.Data.Players))
	}
}

func TestStartFlow(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	readUntil(t, host, ws.MsgRoomState)

	// Starting below minPlayers is refused, with a usable message.
	send(t, host, ws.MsgStart)
	if msg := awaitError(t, host); !strings.Contains(msg, "players") {
		t.Fatalf("error message = %q, want one about player count", msg)
	}

	// Joining is all it takes — the room becomes startable for everyone.
	guest := f.dial(t, rm.Code(), "p2", "Two")
	readUntil(t, guest, ws.MsgRoomState)
	awaitState(t, host, "the room is startable", func(v roomView) bool { return v.CanStart })

	// A non-host cannot start.
	send(t, guest, ws.MsgStart)
	if msg := awaitError(t, guest); !strings.Contains(msg, "host") {
		t.Fatalf("error message = %q, want one about not being host", msg)
	}

	// The host starts, and both players are told.
	send(t, host, ws.MsgStart)
	for name, c := range map[string]*gorilla.Conn{"host": host, "guest": guest} {
		started := readUntil(t, c, ws.MsgGameStarted)
		if started.Data.Status != string(room.StatusInGame) {
			t.Errorf("%s saw status %q, want in_game", name, started.Data.Status)
		}
	}
}

func TestDisconnectHoldsSeat(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	guest := f.dial(t, rm.Code(), "p2", "Two")
	readUntil(t, host, ws.MsgRoomState)
	readUntil(t, guest, ws.MsgRoomState)

	guest.Close()

	view := awaitState(t, host, "the dropped player is marked disconnected", func(v roomView) bool {
		present, connected := playerByID(v, "p2")
		return present && !connected
	})
	if len(view.Data.Players) != 2 {
		t.Errorf("players = %d, want the dropped player's seat to be held", len(view.Data.Players))
	}
}

func TestLeaveRemovesPlayer(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	guest := f.dial(t, rm.Code(), "p2", "Two")
	readUntil(t, host, ws.MsgRoomState)
	readUntil(t, guest, ws.MsgRoomState)

	send(t, guest, ws.MsgLeave)

	view := awaitState(t, host, "the leaving player is gone", func(v roomView) bool {
		present, _ := playerByID(v, "p2")
		return !present
	})
	if len(view.Data.Players) != 1 || view.Data.Players[0].ID != "p1" {
		t.Errorf("players = %+v, want only p1 remaining", view.Data.Players)
	}
}

func TestUnknownMessageType(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	readUntil(t, host, ws.MsgRoomState)

	send(t, host, "definitely-not-a-real-type")

	if msg := awaitError(t, host); !strings.Contains(msg, "unknown message type") {
		t.Errorf("error message = %q, want one naming the unknown type", msg)
	}
}

func TestReconnectReclaimsSeat(t *testing.T) {
	f := newFixture(t, 2, 4)
	rm, _ := f.rooms.Create("test-game", room.Private)

	host := f.dial(t, rm.Code(), "p1", "One")
	guest := f.dial(t, rm.Code(), "p2", "Two")
	readUntil(t, host, ws.MsgRoomState)
	readUntil(t, guest, ws.MsgRoomState)

	guest.Close()
	awaitState(t, host, "the guest is marked disconnected", func(v roomView) bool {
		present, connected := playerByID(v, "p2")
		return present && !connected
	})

	// Same player id — the seat is reclaimed rather than duplicated.
	f.dial(t, rm.Code(), "p2", "Two")

	view := awaitState(t, host, "the guest is connected again", func(v roomView) bool {
		present, connected := playerByID(v, "p2")
		return present && connected
	})
	if len(view.Data.Players) != 2 {
		t.Errorf("players = %d, want 2 after reconnect (no duplicate seat)", len(view.Data.Players))
	}
}

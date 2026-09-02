package ws

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/room"
)

// Handler upgrades room connections and routes client messages to the room.
type Handler struct {
	auth     *auth.Service
	rooms    *room.Manager
	registry *Registry
	upgrader websocket.Upgrader
}

// NewHandler wires the socket handler to its dependencies. allowedOrigin is
// matched against the request's Origin header; "*" disables the check, which
// is fine in dev and for a no-cookie API.
func NewHandler(authService *auth.Service, rooms *room.Manager, registry *Registry, allowedOrigin string) *Handler {
	return &Handler{
		auth:     authService,
		rooms:    rooms,
		registry: registry,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if allowedOrigin == "*" {
					return true
				}
				return r.Header.Get("Origin") == allowedOrigin
			},
		},
	}
}

// ServeHTTP handles GET /api/ws?code=ABC123&token=... The token travels in the
// query string because the browser WebSocket API cannot set headers; it is
// verified exactly like the HTTP Authorization header would be.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, err := h.auth.Verify(r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, "invalid or expired session", http.StatusUnauthorized)
		return
	}

	rm, err := h.rooms.Get(r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	// Joining here as well as over HTTP keeps the socket the single source of
	// truth for membership: a shared link opened directly still works, and a
	// reconnect reclaims the same seat by player id.
	if err := rm.Join(claims.PlayerID, claims.Nickname); err != nil {
		status := http.StatusConflict
		if errors.Is(err, room.ErrRoomFull) {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}

	socket, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade already wrote a response.
	}

	c := newConn(socket, claims.PlayerID, claims.Nickname)
	hub := h.registry.hubFor(rm)
	hub.add(c)
	rm.SetConnected(claims.PlayerID, true)

	go c.writePump()
	hub.broadcastState(MsgRoomState)

	c.readMessages(func(msg Inbound) { h.handleMessage(rm, hub, c, msg) })

	// Socket gone. Hold the seat if the player has no other tab open — the
	// manager's reaper frees it once the grace period passes.
	c.close()
	if stillConnected := hub.remove(c); !stillConnected {
		rm.SetConnected(claims.PlayerID, false)
	}
	hub.broadcastState(MsgRoomState)
	h.registry.dropIfEmpty(rm.Code())
}

func (h *Handler) handleMessage(rm *room.Room, hub *hub, c *conn, msg Inbound) {
	switch msg.Type {
	case MsgPing:
		c.send(Outbound{Type: MsgPong})
		return

	case MsgStart:
		if err := rm.Start(c.playerID); err != nil {
			c.send(errorMessage(startErrorMessage(err)))
			return
		}
		hub.broadcastState(MsgGameStarted)
		return

	case MsgLeave:
		rm.Leave(c.playerID)
		c.close()

	default:
		c.send(errorMessage("unknown message type: " + msg.Type))
		return
	}

	hub.broadcastState(MsgRoomState)
}

// startErrorMessage turns start failures into something a player can act on.
func startErrorMessage(err error) string {
	switch {
	case errors.Is(err, room.ErrNotHost):
		return "only the host can start the game"
	case errors.Is(err, room.ErrNotEnough):
		return "not enough players yet"
	case errors.Is(err, room.ErrGameInProgress):
		return "the game has already started"
	default:
		log.Printf("ws: unexpected start error: %v", err)
		return "could not start the game"
	}
}

package ws

import (
	"sync"

	"github.com/kaviraj-j/playard/server/internal/room"
)

// hub fans messages out to every socket in one room. Each player may have
// more than one socket (a second tab), so connections are keyed by socket.
type hub struct {
	mu    sync.RWMutex
	conns map[*conn]struct{}
	room  *room.Room
}

func newHub(r *room.Room) *hub {
	return &hub{conns: make(map[*conn]struct{}), room: r}
}

func (h *hub) add(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[c] = struct{}{}
}

// remove drops a socket, reporting whether the player has any socket left.
func (h *hub) remove(c *conn) (playerStillConnected bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.conns, c)
	for other := range h.conns {
		if other.playerID == c.playerID {
			return true
		}
	}
	return false
}

func (h *hub) isEmpty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns) == 0
}

// broadcastState sends every socket the room view rendered for its own
// player, so per-player hidden state stays hidden.
func (h *hub) broadcastState(msgType string) {
	h.mu.RLock()
	conns := make([]*conn, 0, len(h.conns))
	for c := range h.conns {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		c.send(Outbound{Type: msgType, Data: h.room.ViewFor(c.playerID)})
	}
}

// Registry maps room codes to their hubs.
type Registry struct {
	mu   sync.RWMutex
	hubs map[string]*hub
}

// NewRegistry creates an empty hub registry.
func NewRegistry() *Registry {
	return &Registry{hubs: make(map[string]*hub)}
}

func (reg *Registry) hubFor(r *room.Room) *hub {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	h, ok := reg.hubs[r.Code()]
	if !ok {
		h = newHub(r)
		reg.hubs[r.Code()] = h
	}
	return h
}

func (reg *Registry) dropIfEmpty(code string) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	if h, ok := reg.hubs[code]; ok && h.isEmpty() {
		delete(reg.hubs, code)
	}
}

// Broadcast pushes the current room state to everyone in the room, if anyone
// is connected. It is the hook the room manager's reaper calls.
func (reg *Registry) Broadcast(r *room.Room) {
	reg.mu.RLock()
	h, ok := reg.hubs[r.Code()]
	reg.mu.RUnlock()
	if ok {
		h.broadcastState(MsgRoomState)
	}
}

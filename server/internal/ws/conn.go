package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4 << 10
	sendBuffer     = 16
)

// conn is one player's socket. Writes go through the send channel so only the
// write pump ever touches the underlying websocket.
type conn struct {
	ws       *websocket.Conn
	playerID string
	nickname string
	out      chan Outbound
	closed   chan struct{}
}

func newConn(socket *websocket.Conn, playerID, nickname string) *conn {
	return &conn{
		ws:       socket,
		playerID: playerID,
		nickname: nickname,
		out:      make(chan Outbound, sendBuffer),
		closed:   make(chan struct{}),
	}
}

// send queues a message, dropping it if the client is too slow rather than
// blocking the broadcaster on one stalled socket.
func (c *conn) send(msg Outbound) {
	select {
	case c.out <- msg:
	case <-c.closed:
	default:
		log.Printf("ws: dropping message for slow client %s", c.playerID)
	}
}

func (c *conn) close() {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
}

// writePump owns all writes to the socket, including keepalive pings.
func (c *conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg := <-c.out:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.closed:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

// readMessages yields decoded client messages until the socket closes.
func (c *conn) readMessages(handle func(Inbound)) {
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.ws.ReadMessage()
		if err != nil {
			return
		}

		var msg Inbound
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.send(errorMessage("malformed message"))
			continue
		}
		handle(msg)
	}
}

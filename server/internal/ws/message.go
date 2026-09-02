// Package ws carries the live room connection: one WebSocket per player,
// grouped into a per-room hub that broadcasts state changes.
package ws

import "encoding/json"

// Inbound message types, client → server.
const (
	MsgStart = "start"
	MsgLeave = "leave"
	MsgPing  = "ping"
)

// Outbound message types, server → client.
const (
	MsgRoomState   = "room_state"
	MsgGameStarted = "game_started"
	MsgError       = "error"
	MsgPong        = "pong"
)

// Inbound is a message from a client. Payload is game-specific and only
// decoded by the game implementation.
type Inbound struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Outbound is a message to a client.
type Outbound struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func errorMessage(message string) Outbound {
	return Outbound{Type: MsgError, Data: map[string]string{"message": message}}
}

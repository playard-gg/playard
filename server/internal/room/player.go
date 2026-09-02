package room

import "time"

// Player is a room member. Identity comes from the session token, so a player
// who reconnects with the same token reclaims the same seat.
type Player struct {
	ID       string
	Nickname string
	// Connected is false while the player's socket is gone but their seat is
	// still being held.
	Connected      bool
	DisconnectedAt time.Time
}

// PlayerView is the wire representation of a player.
type PlayerView struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	Connected bool   `json:"connected"`
	IsHost    bool   `json:"is_host"`
}

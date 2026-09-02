// Package httpapi exposes the request/response endpoints that sit in front of
// the live socket: the game catalog, room creation, join, and quickmatch.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/room"
)

// API holds the handlers' dependencies, injected once in main.
type API struct {
	games *games.Registry
	rooms *room.Manager
}

// New constructs the HTTP API.
func New(registry *games.Registry, rooms *room.Manager) *API {
	return &API{games: registry, rooms: rooms}
}

// ListGames serves the game catalog. It needs no auth — it is the same for
// everyone and drives the pre-login marketing view too.
func (a *API) ListGames(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"games": a.games.All()})
}

type createRoomRequest struct {
	GameID     string          `json:"game_id"`
	Visibility room.Visibility `json:"visibility"`
}

type roomRefResponse struct {
	Code   string `json:"code"`
	GameID string `json:"game_id"`
}

// CreateRoom makes a room for a game and returns its join code. The caller
// does not become a member here — that happens when their socket connects.
func (a *API) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	visibility := req.Visibility
	if visibility != room.Public {
		visibility = room.Private
	}

	rm, err := a.rooms.Create(req.GameID, visibility)
	if err != nil {
		if errors.Is(err, games.ErrUnknownGame) {
			writeError(w, http.StatusNotFound, "unknown game")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create room")
		return
	}

	writeJSON(w, http.StatusCreated, roomRefResponse{Code: rm.Code(), GameID: rm.GameID()})
}

type joinRoomRequest struct {
	Code string `json:"code"`
}

// JoinRoom validates that a code refers to a joinable room. It is a
// pre-flight check so the client can show a useful error before opening a
// socket; the socket handler performs the real join.
func (a *API) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rm, err := a.rooms.Get(req.Code)
	if err != nil {
		writeError(w, http.StatusNotFound, "no room with that code")
		return
	}

	summary := rm.Summarize()
	if summary.Status != room.StatusLobby {
		writeError(w, http.StatusConflict, "that game has already started")
		return
	}
	if summary.PlayerCount >= summary.MaxPlayers {
		writeError(w, http.StatusForbidden, "that room is full")
		return
	}

	writeJSON(w, http.StatusOK, roomRefResponse{Code: rm.Code(), GameID: rm.GameID()})
}

type quickmatchRequest struct {
	GameID string `json:"game_id"`
}

// Quickmatch drops the caller into an open public room for the game, creating
// one if none is waiting.
func (a *API) Quickmatch(w http.ResponseWriter, r *http.Request) {
	var req quickmatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rm, err := a.rooms.Quickmatch(req.GameID)
	if err != nil {
		if errors.Is(err, games.ErrUnknownGame) {
			writeError(w, http.StatusNotFound, "unknown game")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not find a match")
		return
	}

	writeJSON(w, http.StatusOK, roomRefResponse{Code: rm.Code(), GameID: rm.GameID()})
}

// GetRoom returns a pre-join preview so a shared link can show what it leads
// to before the player commits.
func (a *API) GetRoom(w http.ResponseWriter, r *http.Request) {
	rm, err := a.rooms.Get(r.PathValue("code"))
	if err != nil {
		writeError(w, http.StatusNotFound, "no room with that code")
		return
	}
	writeJSON(w, http.StatusOK, rm.Summarize())
}

// Routes registers every endpoint on the mux, applying auth and the rate
// limiter to the room endpoints only.
func (a *API) Routes(mux *http.ServeMux, authService *auth.Service, limit func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/games", a.ListGames)

	guard := func(h http.HandlerFunc) http.HandlerFunc {
		return limit(authService.RequireAuth(h))
	}
	mux.HandleFunc("GET /api/rooms/{code}", guard(a.GetRoom))
	mux.HandleFunc("POST /api/rooms", guard(a.CreateRoom))
	mux.HandleFunc("POST /api/rooms/join", guard(a.JoinRoom))
	mux.HandleFunc("POST /api/rooms/quickmatch", guard(a.Quickmatch))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

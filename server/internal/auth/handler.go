package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	minNicknameLen = 2
	maxNicknameLen = 20
)

type loginRequest struct {
	Nickname string `json:"nickname"`
}

type loginResponse struct {
	Token    string `json:"token"`
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
}

// LoginHandler issues a signed session token for a chosen nickname.
func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	nickname := strings.TrimSpace(req.Nickname)
	if len(nickname) < minNicknameLen || len(nickname) > maxNicknameLen {
		writeError(w, http.StatusBadRequest, "nickname must be between 2 and 20 characters")
		return
	}

	playerID, err := NewPlayerID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	token, err := s.Issue(playerID, nickname)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{
		Token:    token,
		PlayerID: playerID,
		Nickname: nickname,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

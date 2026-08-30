package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const tokenTTL = 7 * 24 * time.Hour

var (
	ErrMalformedToken = errors.New("auth: malformed token")
	ErrInvalidToken   = errors.New("auth: invalid token signature")
	ErrExpiredToken   = errors.New("auth: token expired")
)

// Claims is the signed payload identifying a player.
type Claims struct {
	PlayerID  string `json:"player_id"`
	Nickname  string `json:"nickname"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

// NewPlayerID generates a random UUIDv4 player identifier.
func NewPlayerID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("auth: generate player id: %w", err)
	}
	return id.String(), nil
}

// Service issues and verifies signed session tokens using an injected
// secret, rather than reading it from the environment itself.
type Service struct {
	secret []byte
}

// NewService constructs a token Service with the given HMAC secret.
func NewService(secret []byte) *Service {
	return &Service{secret: secret}
}

// Issue signs a new token for the given player.
func (s *Service) Issue(playerID, nickname string) (string, error) {
	now := time.Now()
	claims := Claims{
		PlayerID:  playerID,
		Nickname:  nickname,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(tokenTTL).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("auth: marshal claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := s.sign(encodedPayload)

	return encodedPayload + "." + signature, nil
}

// Verify checks a token's signature and expiry, returning its claims.
func (s *Service) Verify(token string) (Claims, error) {
	encodedPayload, signature, found := strings.Cut(token, ".")
	if !found {
		return Claims{}, ErrMalformedToken
	}

	expectedSignature := s.sign(encodedPayload)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
		return Claims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return Claims{}, ErrMalformedToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrMalformedToken
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}

func (s *Service) sign(encodedPayload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(encodedPayload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

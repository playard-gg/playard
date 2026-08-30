package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	svc := NewService([]byte("test-secret"))

	playerID, err := NewPlayerID()
	if err != nil {
		t.Fatalf("NewPlayerID() error = %v", err)
	}

	token, err := svc.Issue(playerID, "kavi")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.PlayerID != playerID {
		t.Errorf("PlayerID = %q, want %q", claims.PlayerID, playerID)
	}
	if claims.Nickname != "kavi" {
		t.Errorf("Nickname = %q, want %q", claims.Nickname, "kavi")
	}
}

func TestVerify(t *testing.T) {
	svc := NewService([]byte("test-secret"))

	validToken, err := svc.Issue("player-1", "kavi")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	expiredClaims := Claims{
		PlayerID:  "player-1",
		Nickname:  "kavi",
		IssuedAt:  time.Now().Add(-2 * tokenTTL).Unix(),
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	}
	expiredToken := signClaims(t, svc, expiredClaims)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{name: "valid token", token: validToken, wantErr: nil},
		{name: "malformed token", token: "not-a-real-token", wantErr: ErrMalformedToken},
		{name: "tampered signature", token: validToken[:len(validToken)-1] + "x", wantErr: ErrInvalidToken},
		{name: "expired token", token: expiredToken, wantErr: ErrExpiredToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Verify(tt.token)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("Verify() unexpected error = %v", err)
			}
			if tt.wantErr != nil && err != tt.wantErr {
				t.Fatalf("Verify() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	issuer := NewService([]byte("secret-a"))
	verifier := NewService([]byte("secret-b"))

	token, err := issuer.Issue("player-1", "kavi")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if _, err := verifier.Verify(token); err != ErrInvalidToken {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidToken)
	}
}

func signClaims(t *testing.T, svc *Service, claims Claims) string {
	t.Helper()
	raw, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	return payload + "." + svc.sign(payload)
}

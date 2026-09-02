package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaviraj-j/playard/server/internal/ratelimit"
)

func TestAllow(t *testing.T) {
	l := ratelimit.New(1, 3)

	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d was blocked inside the burst allowance", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Error("request past the burst allowance was permitted")
	}
	if !l.Allow("5.6.7.8") {
		t.Error("a different caller was blocked by another caller's usage")
	}
}

func TestMiddleware(t *testing.T) {
	l := ratelimit.New(1, 1)
	handler := l.Middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	call := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/rooms", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec.Code
	}

	if got := call(); got != http.StatusOK {
		t.Fatalf("first call status = %d, want 200", got)
	}
	if got := call(); got != http.StatusTooManyRequests {
		t.Errorf("second call status = %d, want 429", got)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{"remote addr", "1.2.3.4:5678", "", "1.2.3.4"},
		{"single forwarded", "10.0.0.1:80", "1.2.3.4", "1.2.3.4"},
		{"forwarded chain uses the client", "10.0.0.1:80", "1.2.3.4, 10.0.0.2", "1.2.3.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if got := ratelimit.ClientIP(req); got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

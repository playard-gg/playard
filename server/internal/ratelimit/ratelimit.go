// Package ratelimit provides a small per-IP token bucket. The goal is to stop
// an accidental or casual flood of room creates and code-guessing joins — not
// to be a full abuse system.
package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// Limiter refills each caller's bucket at rate tokens/second up to burst.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
}

// New builds a limiter allowing burst requests immediately, refilling at rate
// per second.
func New(rate, burst float64) *Limiter {
	return &Limiter{buckets: make(map[string]*bucket), rate: rate, burst: burst}
}

// Allow consumes a token for key, reporting whether the request may proceed.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.burst - 1, lastSeen: now}
		return true
	}

	b.tokens = min(l.burst, b.tokens+now.Sub(b.lastSeen).Seconds()*l.rate)
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware rejects requests from callers over the limit with 429.
func (l *Limiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(ClientIP(r)) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests, slow down"}`))
			return
		}
		next(w, r)
	}
}

// StartCleanup drops idle buckets so the map does not grow without bound.
func (l *Limiter) StartCleanup(done <-chan struct{}, idle time.Duration) {
	go func() {
		ticker := time.NewTicker(idle)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case now := <-ticker.C:
				l.mu.Lock()
				for key, b := range l.buckets {
					if now.Sub(b.lastSeen) > idle {
						delete(l.buckets, key)
					}
				}
				l.mu.Unlock()
			}
		}
	}()
}

// ClientIP resolves the caller's address, honouring X-Forwarded-For so the
// limiter still works behind a single trusted proxy.
func ClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if idx := len(fwd); idx > 0 {
			for i, c := range fwd {
				if c == ',' {
					return fwd[:i]
				}
			}
			return fwd
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

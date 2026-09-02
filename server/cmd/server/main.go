package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/config"
	"github.com/kaviraj-j/playard/server/internal/games"
	"github.com/kaviraj-j/playard/server/internal/games/catalog"
	"github.com/kaviraj-j/playard/server/internal/httpapi"
	"github.com/kaviraj-j/playard/server/internal/ratelimit"
	"github.com/kaviraj-j/playard/server/internal/room"
	"github.com/kaviraj-j/playard/server/internal/ws"
)

const (
	// roomRate/roomBurst let a player create or join freely in normal use
	// while capping code-guessing to a few attempts per second per IP.
	roomRate  = 1.0
	roomBurst = 10.0

	bucketIdle = 10 * time.Minute
)

type healthResponse struct {
	Status    string `json:"status"`
	UptimeSec int64  `json:"uptime_seconds"`
	Time      string `json:"time"`
}

func main() {
	startedAt := time.Now()
	cfg := config.Load()

	registry, err := games.NewRegistry(catalog.All()...)
	if err != nil {
		log.Fatalf("game catalog is invalid: %v", err)
	}

	authService := auth.NewService(cfg.AuthSecret)
	rooms := room.NewManager(registry, room.DefaultGrace)
	hubs := ws.NewRegistry()
	// The reaper frees seats and rooms in the background; hubs rebroadcast so
	// remaining players see someone leave without waiting for another event.
	rooms.OnChange(hubs.Broadcast)

	limiter := ratelimit.New(roomRate, roomBurst)

	done := make(chan struct{})
	defer close(done)
	rooms.StartReaper(done)
	limiter.StartCleanup(done, bucketIdle)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Status:    "ok",
			UptimeSec: int64(time.Since(startedAt).Seconds()),
			Time:      time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("POST /api/auth/login", authService.LoginHandler)
	httpapi.New(registry, rooms).Routes(mux, authService, limiter.Middleware)
	mux.Handle("GET /api/ws", ws.NewHandler(authService, rooms, hubs, cfg.CORSOrigin))

	log.Printf("playard server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, withCORS(mux, cfg.CORSOrigin)); err != nil {
		log.Fatal(err)
	}
}

// withCORS allows the configured origin (the Vite dev server by default) to
// call the API. Playard has no cookie-based auth, so a permissive origin
// carries no CSRF risk.
func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kaviraj-j/playard/server/internal/auth"
	"github.com/kaviraj-j/playard/server/internal/config"
)

type healthResponse struct {
	Status    string `json:"status"`
	UptimeSec int64  `json:"uptime_seconds"`
	Time      string `json:"time"`
}

func main() {
	startedAt := time.Now()
	cfg := config.Load()

	authService := auth.NewService(cfg.AuthSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Status:    "ok",
			UptimeSec: int64(time.Since(startedAt).Seconds()),
			Time:      time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/auth/login", authService.LoginHandler)

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

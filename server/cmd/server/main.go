package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string `json:"status"`
	UptimeSec int64  `json:"uptime_seconds"`
	Time      string `json:"time"`
}

func main() {
	startedAt := time.Now()

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

	addr := ":8080"
	log.Printf("playard server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

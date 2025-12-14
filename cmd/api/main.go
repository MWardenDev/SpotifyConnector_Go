package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MWardenDev/SpotifyConnector_Go/internal/config"
	"github.com/MWardenDev/SpotifyConnector_Go/internal/handlers"
	"github.com/MWardenDev/SpotifyConnector_Go/internal/spotify"	
	"github.com/MWardenDev/SpotifyConnector_Go/internal/store"

)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	st := store.NewMemoryTokenStore()
	sp := spotify.NewClient(cfg)
	h := handlers.New(sp, st)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /auth/login", h.AuthLogin)
	mux.HandleFunc("GET /auth/callback", h.AuthCallback)
	mux.HandleFunc("GET /me", h.Me)

	addr := ":" + cfg.Port
	fmt.Printf("SpotifyConnector Go API listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

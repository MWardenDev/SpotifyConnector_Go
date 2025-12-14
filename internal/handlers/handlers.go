package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
	
	"spotifyconnector_go/internal/spotify"
	"spotifyconnector_go/internal/store"
)

type Handler struct {
	Spotify 	*spotify.Client
	Strore		store.TokenStore
}

func New(sp *spotify.Client, st store.Token) *Handler {
	return &Handler{Spotify: sp, Store: st}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) AuthLogin(w http.ResponseWriter, r *http.Request) {
	state := randomHex(16)

	http.SetCookie(w, &http.Cookie{
		Name:     "sc_oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, h.Spotify.AuthorizeURL(state), http.StatusFound)
}
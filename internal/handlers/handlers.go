package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MWardenDev/SpotifyConnector_Go/internal/spotify"
	"github.com/MWardenDev/SpotifyConnector_Go/internal/store"
)

type Handler struct {
	Spotify *spotify.Client
	Store   store.TokenStore
}

func New(sp *spotify.Client, st store.TokenStore) *Handler {
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

func (h *Handler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")

	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	c, err := r.Cookie("sc_oauth_state")
	if err != nil || c.Value != state {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	token, err := h.Spotify.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	sessionID := randomHex(24)
	h.Store.Put(sessionID, token)

	http.SetCookie(w, &http.Cookie{
		Name:     "sc_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// clear oauth state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sc_oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/me", http.StatusFound)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sc_session")
	if err != nil || c.Value == "" {
		http.Error(w, "not authenticated. visit /auth/login", http.StatusUnauthorized)
		return
	}

	tok, ok := h.Store.Get(c.Value)
	if !ok || tok.AccessToken == "" {
		http.Error(w, "session not found. visit /auth/login", http.StatusUnauthorized)
		return
	}

	me, err := h.Spotify.GetMe(r.Context(), tok.AccessToken)
	if err != nil {
		http.Error(w, "spotify /me failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, me)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

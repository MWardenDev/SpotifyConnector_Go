package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port			string
	ClientID		string
	ClientSecret	string
	RedirectURI		string
	Scopes			string
	PublicBaseURL	string
}

func Load() (Config, error) {
	cfg := Config{
		Port:		envOr("Port", "5001"),
		ClientID: 	os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURI: os.Getenv("SPOTIFY_REDIRECT_URI"),
		Scopes: envOr("SPOTIFY_SCOPES", "user-read-email user-read-private"),
		PublicBaseURL: envOr("PUBLIC_BASE_URL", ""),
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return Config{}, fmt.Errorf("missing required env vars: SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, SPOTIFY_REDIRECT_URI")
	}

	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
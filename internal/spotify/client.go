package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MWardenDev/SpotifyConnector_Go/internal/config"
	"github.com/MWardenDev/SpotifyConnector_Go/internal/store"
)

type Client struct {
	cfg config.Config
	http *http.Client
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) AuthorizeURL(state string) string {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", c.cfg.ClientID)
	v.Set("redirect_uri", c.cfg.RedirectURI)
	v.Set("scope", c.cfg.Scopes)
	v.Set("state", state)
	v.Set("show_dialog", "false")

	return "https://accounts.spotify.com/authorize?" + v.Encode()
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (store.Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.cfg.RedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return store.Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Basic auth: base64(client_id:client_secret)
	basic := base64.StdEncoding.EncodeToString([]byte(c.cfg.ClientID + ":" + c.cfg.ClientSecret))
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := c.http.Do(req)
	if err != nil {
		return store.Token{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return store.Token{}, fmt.Errorf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tr struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return store.Token{}, err
	}

	return store.Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
		ExpiresIn:    tr.ExpiresIn,
		Scope:        tr.Scope,
	}, nil
}

func (c *Client) GetMe(ctx context.Context, accessToken string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("me failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

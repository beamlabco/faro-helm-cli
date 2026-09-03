package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/beamlabco/faro-helm-cli/internal/config"
)

// Client is the API client for Faro Helm backend
type Client struct {
	baseURL string
	token   string
	http    *resty.Client

	// cfg is nil for a bare NewClient. Only a client backed by a persisted
	// config (NewClientFromConfig) can refresh its access token — it reads
	// the refresh token from, and persists new tokens back to, cfg.
	cfg *config.Config

	refreshMu    sync.Mutex
	refreshingCh chan struct{}
	refreshErr   error
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")

	if token != "" {
		client.SetAuthToken(token)
	}

	c := &Client{
		baseURL: baseURL,
		token:   token,
		http:    client,
	}

	// Access tokens are short-lived (15m). On a 401, attempt a single
	// refresh-and-retry before giving up. resty re-reads client.Token fresh
	// on every attempt (see its addCredentials middleware), so updating the
	// token inside the retry condition is enough for the retried attempt to
	// pick it up.
	client.SetRetryCount(1)
	client.AddRetryCondition(c.refreshAndRetry)

	return c
}

// NewClientFromConfig creates a new API client from config
func NewClientFromConfig(cfg *config.Config) *Client {
	baseURL := "http://localhost:3001"
	if cfg.API != nil && cfg.API.BaseURL != "" {
		baseURL = cfg.API.BaseURL
	}

	token := ""
	if cfg.Auth != nil {
		token = cfg.Auth.Token
	}

	c := NewClient(baseURL, token)
	c.cfg = cfg
	return c
}

// SetToken updates the client's authentication token
func (c *Client) SetToken(token string) {
	c.token = token
	c.http.SetAuthToken(token)
}

// refreshAndRetry is a resty retry condition: on a 401 it attempts to
// refresh the access token, asking resty to retry the request only if that
// succeeded.
func (c *Client) refreshAndRetry(resp *resty.Response, err error) bool {
	if err != nil || resp == nil || resp.StatusCode() != http.StatusUnauthorized {
		return false
	}
	return c.refreshAccessToken()
}

// refreshAccessToken performs a single-flight token refresh: concurrent
// callers (e.g. the dashboard's parallel fetches on startup) share one
// in-flight refresh instead of each spending the single-use refresh token,
// which would fail every attempt after the first.
func (c *Client) refreshAccessToken() bool {
	c.refreshMu.Lock()
	if ch := c.refreshingCh; ch != nil {
		c.refreshMu.Unlock()
		<-ch
		c.refreshMu.Lock()
		ok := c.refreshErr == nil
		c.refreshMu.Unlock()
		return ok
	}

	ch := make(chan struct{})
	c.refreshingCh = ch
	c.refreshMu.Unlock()

	refreshErr := c.doRefresh()

	c.refreshMu.Lock()
	c.refreshErr = refreshErr
	c.refreshingCh = nil
	c.refreshMu.Unlock()
	close(ch)

	return refreshErr == nil
}

// doRefresh calls POST /auth/refresh and, on success, applies and persists
// the new token pair. On an explicit rejection from the server (refresh
// token invalid, expired, or already used), the local session is cleared
// so the user is treated as logged out — IsAuthenticated() will return
// false and every auth-required command will prompt for /login again,
// instead of the CLI silently believing it's still signed in.
//
// A separate resty client is used for this call so it doesn't itself go
// through refreshAndRetry (which would recurse if the refresh call ever
// 401s).
func (c *Client) doRefresh() error {
	if c.cfg == nil || c.cfg.Auth == nil || c.cfg.Auth.RefreshToken == "" {
		return errors.New("no refresh token available")
	}

	var result AuthRefreshResponse
	resp, err := resty.New().
		SetBaseURL(c.baseURL).
		SetHeader("Content-Type", "application/json").
		R().
		SetBody(map[string]string{"refreshToken": c.cfg.Auth.RefreshToken}).
		SetResult(&result).
		Post("/api/v1/auth/refresh")

	if err != nil {
		// Transient/network failure — not a rejection, don't log the user out.
		return err
	}

	if resp.IsError() {
		refreshErr := parseError(resp)
		c.cfg.Clear()
		_ = config.Save(c.cfg)
		c.SetToken("")
		return refreshErr
	}

	c.SetToken(result.AccessToken)
	c.cfg.Auth = &config.Auth{Token: result.AccessToken, RefreshToken: result.RefreshToken}
	_ = config.Save(c.cfg)

	return nil
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// APIError represents an API error
type APIError struct {
	StatusCode int
	Message    string
	Code       string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s (%s)", e.Message, e.Code)
	}
	return e.Message
}

// parseError extracts error information from response
func parseError(resp *resty.Response) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode(),
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status()),
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode(),
		Message:    errResp.Error,
		Code:       errResp.Code,
		Details:    errResp.Details,
	}
}

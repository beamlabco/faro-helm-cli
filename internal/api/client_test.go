package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beamlabco/faro-helm-cli/internal/config"
)

func newTestConfig(t *testing.T, baseURL, accessToken, refreshToken string) *config.Config {
	t.Helper()
	config.SetConfigDirForTesting(t.TempDir())
	t.Cleanup(func() { config.SetConfigDirForTesting("") })

	cfg := &config.Config{
		API: &config.API{BaseURL: baseURL},
	}
	if accessToken != "" || refreshToken != "" {
		cfg.Auth = &config.Auth{Token: accessToken, RefreshToken: refreshToken}
	}
	return cfg
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// TestGetMe_RefreshesExpiredTokenAndRetries verifies a 401 triggers a
// transparent refresh-and-retry, and that the new tokens are persisted.
func TestGetMe_RefreshesExpiredTokenAndRetries(t *testing.T) {
	var meCalls, refreshCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&meCalls, 1)
		if r.Header.Get("Authorization") != "Bearer new-access" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token", "code": "INVALID_TOKEN"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"account":    map[string]string{"id": "acc1", "email": "a@b.com", "name": "A"},
			"workspaces": []any{},
		})
	})
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		writeJSON(w, http.StatusOK, map[string]string{"accessToken": "new-access", "refreshToken": "new-refresh"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := newTestConfig(t, server.URL, "old-access", "old-refresh")
	client := NewClientFromConfig(cfg)

	me, err := client.GetMe()
	if err != nil {
		t.Fatalf("GetMe() returned error: %v", err)
	}
	if me.Account.ID != "acc1" {
		t.Fatalf("unexpected account: %+v", me.Account)
	}

	if got := atomic.LoadInt32(&meCalls); got != 2 {
		t.Errorf("expected /auth/me to be called twice (fail then retry), got %d", got)
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected exactly one refresh call, got %d", got)
	}

	if cfg.Auth.Token != "new-access" || cfg.Auth.RefreshToken != "new-refresh" {
		t.Errorf("expected in-memory config to hold the refreshed tokens, got %+v", cfg.Auth)
	}

	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	if reloaded.Auth == nil || reloaded.Auth.Token != "new-access" || reloaded.Auth.RefreshToken != "new-refresh" {
		t.Errorf("expected refreshed tokens to be persisted to disk, got %+v", reloaded.Auth)
	}
}

// TestGetMe_ConcurrentRequestsShareOneRefresh verifies concurrent 401s
// (e.g. the dashboard's parallel fetches on startup) trigger only one
// refresh call, since a refresh token is single-use.
func TestGetMe_ConcurrentRequestsShareOneRefresh(t *testing.T) {
	var meCalls, refreshCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&meCalls, 1)
		if r.Header.Get("Authorization") != "Bearer new-access" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token", "code": "INVALID_TOKEN"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"account":    map[string]string{"id": "acc1", "email": "a@b.com", "name": "A"},
			"workspaces": []any{},
		})
	})
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		time.Sleep(50 * time.Millisecond) // give concurrent callers time to pile up
		writeJSON(w, http.StatusOK, map[string]string{"accessToken": "new-access", "refreshToken": "new-refresh"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := newTestConfig(t, server.URL, "old-access", "old-refresh")
	client := NewClientFromConfig(cfg)

	const n = 5
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := client.GetMe()
			errs[i] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: GetMe() returned error: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&refreshCalls); got != 1 {
		t.Errorf("expected exactly one refresh call across %d concurrent 401s, got %d", n, got)
	}
}

// TestGetMe_InvalidRefreshTokenClearsSession verifies that when the
// refresh call itself is rejected, the local session is cleared so the
// user is treated as logged out.
func TestGetMe_InvalidRefreshTokenClearsSession(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token", "code": "INVALID_TOKEN"})
	})
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired refresh token", "code": "INVALID_REFRESH_TOKEN"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := newTestConfig(t, server.URL, "old-access", "old-refresh")
	client := NewClientFromConfig(cfg)

	if _, err := client.GetMe(); err == nil {
		t.Fatal("expected GetMe() to return an error")
	}

	if cfg.Auth != nil {
		t.Errorf("expected session to be cleared after an unrecoverable refresh failure, got %+v", cfg.Auth)
	}
	if cfg.IsAuthenticated() {
		t.Error("expected IsAuthenticated() to be false after an unrecoverable refresh failure")
	}

	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	if reloaded.IsAuthenticated() {
		t.Error("expected the cleared session to be persisted to disk")
	}
}

// TestGetMe_NoRefreshTokenPropagatesOriginalError verifies that with no
// refresh token available, no refresh is attempted and the original 401
// is returned as-is.
func TestGetMe_NoRefreshTokenPropagatesOriginalError(t *testing.T) {
	var refreshCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token", "code": "INVALID_TOKEN"})
	})
	mux.HandleFunc("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		writeJSON(w, http.StatusOK, map[string]string{"accessToken": "new-access", "refreshToken": "new-refresh"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cfg := newTestConfig(t, server.URL, "stale-access", "")
	client := NewClientFromConfig(cfg)

	_, err := client.GetMe()
	if err == nil {
		t.Fatal("expected GetMe() to return an error")
	}
	if apiErr, ok := err.(*APIError); !ok || apiErr.Code != "INVALID_TOKEN" {
		t.Errorf("expected the original INVALID_TOKEN error to propagate, got %v", err)
	}

	if got := atomic.LoadInt32(&refreshCalls); got != 0 {
		t.Errorf("expected no refresh attempt without a refresh token, got %d calls", got)
	}
	if cfg.Auth == nil || cfg.Auth.Token != "stale-access" {
		t.Errorf("expected config to be left untouched, got %+v", cfg.Auth)
	}
}

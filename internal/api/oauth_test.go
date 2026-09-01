package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beamlabco/faro-helm-cli/internal/oauthflow"
)

func TestExchangeAuthorizationCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Errorf("path = %s, want /oauth/token", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		want := map[string]string{
			"grant_type":    "authorization_code",
			"code":          "the-code",
			"redirect_uri":  "http://127.0.0.1:1/callback",
			"client_id":     oauthflow.ClientID,
			"code_verifier": "verifier123",
		}
		for k, v := range want {
			if got := r.Form.Get(k); got != v {
				t.Errorf("form[%q] = %q, want %q", k, got, v)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "AT",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "RT",
			"scope":         "helm:read helm:write profile",
		})
	}))
	defer srv.Close()

	client := NewAuthClient(srv.URL)
	result, err := client.ExchangeAuthorizationCode("the-code", "http://127.0.0.1:1/callback", "verifier123")
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error: %v", err)
	}
	if result.AccessToken != "AT" || result.RefreshToken != "RT" || result.Scope != "helm:read helm:write profile" {
		t.Errorf("result = %+v, want AT/RT/helm:read helm:write profile", result)
	}
}

func TestExchangeAuthorizationCode_SurfacesOAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid, expired, or already-used authorization code",
		})
	}))
	defer srv.Close()

	client := NewAuthClient(srv.URL)
	_, err := client.ExchangeAuthorizationCode("bad-code", "http://127.0.0.1:1/callback", "verifier")
	if err == nil {
		t.Fatal("expected an error for a rejected exchange")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error = %q, want it to mention invalid_grant", err.Error())
	}
}

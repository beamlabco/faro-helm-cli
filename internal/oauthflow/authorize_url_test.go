package oauthflow

import (
	"net/url"
	"testing"
)

func TestBuildAuthorizeURL_SetsAllRequiredParams(t *testing.T) {
	raw := BuildAuthorizeURL(AuthorizeURLParams{
		AuthBaseURL:   "https://auth.farohelm.com",
		ClientID:      "faro-helm-cli",
		RedirectURI:   "http://127.0.0.1:53127/callback",
		Scope:         "helm:read helm:write profile",
		State:         "abc123",
		CodeChallenge: "the-challenge",
	})

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("BuildAuthorizeURL produced an unparseable URL: %v", err)
	}

	if got := u.Scheme + "://" + u.Host + u.Path; got != "https://auth.farohelm.com/oauth/authorize" {
		t.Errorf("base URL = %q, want https://auth.farohelm.com/oauth/authorize", got)
	}

	q := u.Query()
	cases := map[string]string{
		"response_type":         "code",
		"client_id":             "faro-helm-cli",
		"redirect_uri":          "http://127.0.0.1:53127/callback",
		"scope":                 "helm:read helm:write profile",
		"state":                 "abc123",
		"code_challenge":        "the-challenge",
		"code_challenge_method": "S256",
	}
	for key, want := range cases {
		if got := q.Get(key); got != want {
			t.Errorf("query param %q = %q, want %q", key, got, want)
		}
	}
}

func TestBuildAuthorizeURL_EscapesSpecialCharacters(t *testing.T) {
	// A redirect_uri with a query-string-hostile character shouldn't corrupt
	// the surrounding query string.
	raw := BuildAuthorizeURL(AuthorizeURLParams{
		AuthBaseURL:   "https://auth.farohelm.com",
		ClientID:      "faro-helm-cli",
		RedirectURI:   "http://127.0.0.1:53127/callback?extra=a&b=c",
		Scope:         "helm:read",
		State:         "s",
		CodeChallenge: "x",
	})

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("BuildAuthorizeURL produced an unparseable URL: %v", err)
	}
	if got := u.Query().Get("redirect_uri"); got != "http://127.0.0.1:53127/callback?extra=a&b=c" {
		t.Errorf("redirect_uri round-tripped as %q, want the original value intact", got)
	}
}

func TestBuildAuthorizeURL_TrimsTrailingSlashOnAuthBaseURL(t *testing.T) {
	raw := BuildAuthorizeURL(AuthorizeURLParams{
		AuthBaseURL:   "https://auth.farohelm.com/",
		ClientID:      "faro-helm-cli",
		RedirectURI:   "http://127.0.0.1:1/callback",
		State:         "s",
		CodeChallenge: "x",
	})
	u, _ := url.Parse(raw)
	if u.Path != "/oauth/authorize" {
		t.Errorf("path = %q, want /oauth/authorize (no double slash)", u.Path)
	}
}

package oauthflow

import (
	"net/url"
	"strings"
)

// AuthorizeURLParams are the parameters needed to build a GET /oauth/authorize
// request URL. Scope is optional — the auth server falls back to the
// client's full allowed_scopes when omitted.
type AuthorizeURLParams struct {
	AuthBaseURL   string
	ClientID      string
	RedirectURI   string
	Scope         string
	State         string
	CodeChallenge string
}

// BuildAuthorizeURL constructs the GET /oauth/authorize URL for the
// Authorization Code + PKCE grant. code_challenge_method is always S256.
func BuildAuthorizeURL(p AuthorizeURLParams) string {
	base := strings.TrimRight(p.AuthBaseURL, "/")

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURI)
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	q.Set("state", p.State)
	q.Set("code_challenge", p.CodeChallenge)
	q.Set("code_challenge_method", "S256")

	return base + "/oauth/authorize?" + q.Encode()
}

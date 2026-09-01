package api

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/beamlabco/faro-helm-cli/internal/oauthflow"
)

// AuthClient calls faro-auth-api.
type AuthClient struct {
	baseURL string
	http    *resty.Client
}

// NewAuthClient creates an auth API client.
func NewAuthClient(baseURL string) *AuthClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")
	return &AuthClient{baseURL: baseURL, http: client}
}

// BaseURL returns the auth server's base URL, needed to build the
// /oauth/authorize URL.
func (c *AuthClient) BaseURL() string {
	return c.baseURL
}

// AuthRegisterRequest is the body for POST /api/v1/auth/register.
type AuthRegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	OrganizationName string `json:"organizationName,omitempty"`
}

// MeResponse is returned by GET /api/v1/auth/me.
type MeResponse struct {
	Account    MeAccount     `json:"account"`
	Workspaces []MeWorkspace `json:"workspaces"`
}

type MeAccount struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type MeWorkspace struct {
	WorkspaceID string `json:"workspaceId"`
	MemberID    string `json:"memberId"`
	Role        string `json:"role"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
}

// TokenResponse mirrors POST /oauth/token's RFC 6749 §5.1 response body.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// oauthErrorResponse mirrors /oauth/token's RFC 6749 §5.2 error body.
type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ExchangeAuthorizationCode calls POST /oauth/token with
// grant_type=authorization_code, completing the Authorization Code + PKCE
// flow. faro-helm-cli is a public client — no client_secret.
func (c *AuthClient) ExchangeAuthorizationCode(code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	var result TokenResponse
	resp, err := c.http.R().
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  redirectURI,
			"client_id":     oauthflow.ClientID,
			"code_verifier": codeVerifier,
		}).
		SetResult(&result).
		Post("/oauth/token")
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	if resp.IsError() {
		return nil, parseOAuthError(resp)
	}
	return &result, nil
}

func parseOAuthError(resp *resty.Response) error {
	var oerr oauthErrorResponse
	if err := json.Unmarshal(resp.Body(), &oerr); err == nil && oerr.Error != "" {
		if oerr.ErrorDescription != "" {
			return fmt.Errorf("%s: %s", oerr.Error, oerr.ErrorDescription)
		}
		return fmt.Errorf("%s", oerr.Error)
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode())
}

// ChangePassword changes the authenticated user's password via POST /auth/password.
func (c *AuthClient) ChangePassword(accessToken, currentPassword, newPassword string) error {
	resp, err := c.http.R().
		SetAuthToken(accessToken).
		SetBody(map[string]string{"currentPassword": currentPassword, "newPassword": newPassword}).
		Post("/api/v1/auth/password")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("failed to change password: HTTP %d", resp.StatusCode())
	}
	return nil
}

// Register creates a new account and workspace via POST /auth/register.
// No token is issued — the user must verify their email before logging in.
func (c *AuthClient) Register(req *AuthRegisterRequest) error {
	resp, err := c.http.R().
		SetBody(req).
		Post("/api/v1/auth/register")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("registration failed: HTTP %d", resp.StatusCode())
	}
	return nil
}

// GetMe fetches the authenticated account and workspace memberships.
func (c *AuthClient) GetMe(token string) (*MeResponse, error) {
	var result MeResponse
	resp, err := c.http.R().
		SetAuthToken(token).
		SetResult(&result).
		Get("/api/v1/auth/me")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to fetch user info: HTTP %d", resp.StatusCode())
	}
	return &result, nil
}

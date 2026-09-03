package api

import (
	"fmt"
)

// AuthLoginRequest is the body for POST /auth/login.
type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthLoginResponse is returned by POST /auth/login.
type AuthLoginResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	Account      MeAccount `json:"account"`
}

// AuthRefreshResponse is returned by POST /auth/refresh.
type AuthRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// AuthRegisterRequest is the body for POST /auth/register.
type AuthRegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	OrganizationName string `json:"organizationName,omitempty"`
}

// MeResponse is returned by GET /auth/me.
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

// Login authenticates with email and password via POST /auth/login.
func (c *Client) Login(req *AuthLoginRequest) (*AuthLoginResponse, error) {
	var result AuthLoginResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/auth/login")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("login failed: HTTP %d", resp.StatusCode())
	}
	return &result, nil
}

// ChangePassword changes the authenticated user's password via POST /auth/password.
func (c *Client) ChangePassword(currentPassword, newPassword string) error {
	resp, err := c.http.R().
		SetBody(map[string]string{"currentPassword": currentPassword, "newPassword": newPassword}).
		Post("/api/v1/auth/password")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return parseError(resp)
	}
	return nil
}

// Register creates a new account and workspace via POST /auth/register.
// No token is issued — the user must verify their email before logging in.
func (c *Client) Register(req *AuthRegisterRequest) error {
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
// Callers should ensure the client's token is set (via SetToken) first.
func (c *Client) GetMe() (*MeResponse, error) {
	var result MeResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/v1/auth/me")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, parseError(resp)
	}
	return &result, nil
}

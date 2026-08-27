package api

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// AuthClient calls faro-auth-api for device flow authentication.
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

// DeviceCodeResponse is returned by POST /auth/device/code.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenResponse is returned on approval by POST /auth/device/token.
type DeviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// DeviceTokenPair holds the access and refresh tokens returned after device flow approval.
type DeviceTokenPair struct {
	AccessToken  string
	RefreshToken string
}

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

// AuthRegisterRequest is the body for POST /auth/register.
type AuthRegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	OrganizationName string `json:"organizationName,omitempty"`
}

// DeviceTokenError is the error body when the device code is still pending.
type DeviceTokenError struct {
	Error string `json:"error"`
}

// MeResponse is returned by GET /auth/me.
type MeResponse struct {
	Account    MeAccount    `json:"account"`
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

// InitiateDeviceFlow starts the device authorization flow.
func (c *AuthClient) InitiateDeviceFlow() (*DeviceCodeResponse, error) {
	var result DeviceCodeResponse
	resp, err := c.http.R().
		SetResult(&result).
		Post("/api/v1/auth/device/code")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to initiate device flow: HTTP %d", resp.StatusCode())
	}
	return &result, nil
}

// PollDeviceToken polls for the device token.
// Returns (*DeviceTokenPair, nil) on approval.
// Returns (nil, err) where err.Error() is "authorization_pending", "slow_down",
// "expired_token", or "access_denied" for known non-fatal/fatal states.
func (c *AuthClient) PollDeviceToken(deviceCode string) (*DeviceTokenPair, error) {
	resp, err := c.http.R().
		SetBody(map[string]string{"device_code": deviceCode}).
		Post("/api/v1/auth/device/token")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 200 {
		var result DeviceTokenResponse
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, fmt.Errorf("failed to parse token response: %w", err)
		}
		return &DeviceTokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		}, nil
	}

	var errResp DeviceTokenError
	if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode())
	}
	return nil, fmt.Errorf("%s", errResp.Error)
}

// Login authenticates with email and password via POST /auth/login.
func (c *AuthClient) Login(req *AuthLoginRequest) (*AuthLoginResponse, error) {
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

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
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
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
		Post("/auth/device/code")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to initiate device flow: HTTP %d", resp.StatusCode())
	}
	return &result, nil
}

// PollDeviceToken polls for the device token.
// Returns (token, nil) on approval.
// Returns ("", err) where err.Error() is "authorization_pending", "slow_down",
// "expired_token", or "access_denied" for known non-fatal/fatal states.
func (c *AuthClient) PollDeviceToken(deviceCode string) (string, error) {
	resp, err := c.http.R().
		SetBody(map[string]string{"device_code": deviceCode}).
		Post("/auth/device/token")
	if err != nil {
		return "", err
	}

	if resp.StatusCode() == 200 {
		var result DeviceTokenResponse
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return "", fmt.Errorf("failed to parse token response: %w", err)
		}
		return result.AccessToken, nil
	}

	var errResp DeviceTokenError
	if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode())
	}
	return "", fmt.Errorf("%s", errResp.Error)
}

// GetMe fetches the authenticated account and workspace memberships.
func (c *AuthClient) GetMe(token string) (*MeResponse, error) {
	var result MeResponse
	resp, err := c.http.R().
		SetAuthToken(token).
		SetResult(&result).
		Get("/auth/me")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to fetch user info: HTTP %d", resp.StatusCode())
	}
	return &result, nil
}

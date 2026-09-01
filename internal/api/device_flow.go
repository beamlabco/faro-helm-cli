package api

import (
	"encoding/json"
	"fmt"
)

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

// DeviceTokenError is the error body when the device code is still pending.
type DeviceTokenError struct {
	Error string `json:"error"`
}

// InitiateDeviceFlow starts the device authorization flow.
func (c *Client) InitiateDeviceFlow() (*DeviceCodeResponse, error) {
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
func (c *Client) PollDeviceToken(deviceCode string) (*DeviceTokenPair, error) {
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

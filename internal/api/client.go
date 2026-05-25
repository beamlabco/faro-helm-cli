package api

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/beamlabco/faro-helm/internal/config"
)

// Client is the API client for Faro Helm backend
type Client struct {
	baseURL string
	token   string
	http    *resty.Client
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")

	if token != "" {
		client.SetAuthToken(token)
	}

	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    client,
	}
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

	return NewClient(baseURL, token)
}

// SetToken updates the client's authentication token
func (c *Client) SetToken(token string) {
	c.token = token
	c.http.SetAuthToken(token)
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

package api

import (
	"fmt"
	"time"
)

// StandupRequest represents the standup submission request
type StandupRequest struct {
	Date      string `json:"date"`
	TeamID    string `json:"teamId"`
	Yesterday string `json:"yesterday,omitempty"`
	Today     string `json:"today,omitempty"`
	Blockers  string `json:"blockers,omitempty"`
}

// StandupResponse represents a standup in API response
type StandupResponse struct {
	ID        string    `json:"id"`
	MemberID  string    `json:"memberId"`
	TeamID    string    `json:"teamId"`
	Date      string    `json:"date"`
	Yesterday *string   `json:"yesterday"`
	Today     *string   `json:"today"`
	Blockers  *string   `json:"blockers"`
	CreatedAt time.Time `json:"createdAt"`
	User      *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user,omitempty"`
	Team *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team,omitempty"`
}

// SubmitStandupResponse represents the submit standup response
type SubmitStandupResponse struct {
	Standup *StandupResponse `json:"standup"`
	Message string           `json:"message,omitempty"`
}

// GetMyStandupsResponse represents the get my standups response
type GetMyStandupsResponse struct {
	Standups []*StandupResponse `json:"standups"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// SubmitStandup submits or updates a standup
func (c *Client) SubmitStandup(req *StandupRequest) (*SubmitStandupResponse, error) {
	var result SubmitStandupResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/standups")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetMyStandups retrieves the current user's standup history.
// period, if non-empty, must be one of "today", "week", or "month".
func (c *Client) GetMyStandups(period string, limit, offset int) (*GetMyStandupsResponse, error) {
	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}
	if period != "" {
		params["period"] = period
	}

	var result GetMyStandupsResponse
	resp, err := c.http.R().
		SetQueryParams(params).
		SetResult(&result).
		Get("/api/v1/standups/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

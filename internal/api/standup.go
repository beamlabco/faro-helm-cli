package api

import (
	"fmt"
	"time"
)

// StandupRequest represents the standup submission request
type StandupRequest struct {
	Date      string `json:"date"`
	ProjectID string `json:"projectId"`
	Yesterday string `json:"yesterday,omitempty"`
	Today     string `json:"today,omitempty"`
	Blockers  string `json:"blockers,omitempty"`
}

// StandupResponse represents a standup in API response
type StandupResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ProjectID string    `json:"projectId"`
	Date      string    `json:"date"`
	Yesterday *string   `json:"yesterday"`
	Today     *string   `json:"today"`
	Blockers  *string   `json:"blockers"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	User      *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user,omitempty"`
	Project *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project,omitempty"`
}

// SubmitStandupResponse represents the submit standup response
type SubmitStandupResponse struct {
	Standup *StandupResponse `json:"standup"`
	Message string           `json:"message,omitempty"`
}

// GetTodayStandupsResponse represents the get today standups response
type GetTodayStandupsResponse struct {
	Standups []*StandupResponse `json:"standups"`
	Date     string             `json:"date"`
}

// GetMyStandupsResponse represents the get my standups response
type GetMyStandupsResponse struct {
	Standups   []*StandupResponse `json:"standups"`
	Pagination *Pagination        `json:"pagination"`
}

// GetByDateStandupsResponse represents the get standups by date response
type GetByDateStandupsResponse struct {
	Standups []*StandupResponse `json:"standups"`
	Date     string             `json:"date"`
}

// Pagination represents pagination information
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// SubmitStandup submits or updates a standup
func (c *Client) SubmitStandup(req *StandupRequest) (*SubmitStandupResponse, error) {
	var result SubmitStandupResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/standups")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetTodayStandups retrieves all team standups for today
func (c *Client) GetTodayStandups() (*GetTodayStandupsResponse, error) {
	var result GetTodayStandupsResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/standups/today")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetMyStandups retrieves the current user's standup history
func (c *Client) GetMyStandups(limit, offset int) (*GetMyStandupsResponse, error) {
	var result GetMyStandupsResponse
	resp, err := c.http.R().
		SetQueryParams(map[string]string{
			"limit":  fmt.Sprintf("%d", limit),
			"offset": fmt.Sprintf("%d", offset),
		}).
		SetResult(&result).
		Get("/api/standups/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetStandupsByDate retrieves team standups for a specific date
func (c *Client) GetStandupsByDate(date string) (*GetByDateStandupsResponse, error) {
	var result GetByDateStandupsResponse
	resp, err := c.http.R().
		SetQueryParam("date", date).
		SetResult(&result).
		Get("/api/standups")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

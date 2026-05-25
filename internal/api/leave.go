package api

import (
	"fmt"
	"time"
)

// CreateLeaveRequest represents the leave creation request
type CreateLeaveRequest struct {
	Type      string `json:"type"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Reason    string `json:"reason,omitempty"`
}

// LeaveResponse represents a leave record in API response
type LeaveResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Type       string    `json:"type"`
	StartDate  string    `json:"startDate"`
	EndDate    string    `json:"endDate"`
	Reason     *string   `json:"reason"`
	Status     string    `json:"status"`
	ReviewedBy *string   `json:"reviewedBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	User       *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user,omitempty"`
	Reviewer *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"reviewer,omitempty"`
}

// GetLeavesResponse represents the get leaves response
type GetLeavesResponse struct {
	Leaves []*LeaveResponse `json:"leaves"`
	Count  int              `json:"count"`
}

// UpdateLeaveStatusRequest represents the leave status update request
type UpdateLeaveStatusRequest struct {
	Status string `json:"status"`
}

// CreateLeave creates a new leave request
func (c *Client) CreateLeave(req *CreateLeaveRequest) (*LeaveResponse, error) {
	var result LeaveResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/leaves")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetLeaves retrieves leaves with optional filters
func (c *Client) GetLeaves(status string, userID string, limit, offset int) (*GetLeavesResponse, error) {
	var result GetLeavesResponse
	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}
	if status != "" {
		params["status"] = status
	}
	if userID != "" {
		params["userId"] = userID
	}

	resp, err := c.http.R().
		SetQueryParams(params).
		SetResult(&result).
		Get("/api/leaves")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// UpdateLeaveStatus approves or rejects a leave request
func (c *Client) UpdateLeaveStatus(leaveID string, status string) (*LeaveResponse, error) {
	var result LeaveResponse
	resp, err := c.http.R().
		SetBody(&UpdateLeaveStatusRequest{Status: status}).
		SetResult(&result).
		Patch(fmt.Sprintf("/api/leaves/%s", leaveID))

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// CancelLeave cancels a pending leave request
func (c *Client) CancelLeave(leaveID string) error {
	resp, err := c.http.R().
		Delete(fmt.Sprintf("/api/leaves/%s", leaveID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return parseError(resp)
	}

	return nil
}

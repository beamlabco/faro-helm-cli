package api

import (
	"fmt"
	"time"
)

// CreateLeaveRequest represents the leave creation request
type CreateLeaveRequest struct {
	TypeID    string `json:"typeId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Reason    string `json:"reason,omitempty"`
}

// LeaveTypeRef is the leave type nested inside a leave list/history record.
type LeaveTypeRef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ResetPolicy string `json:"resetPolicy"`
}

// LeaveResponse represents a leave record as returned by the list/history/upcoming/pending endpoints.
type LeaveResponse struct {
	ID         string        `json:"id"`
	MemberID   string        `json:"memberId"`
	Type       *LeaveTypeRef `json:"type"`
	StartDate  string        `json:"startDate"`
	EndDate    string        `json:"endDate"`
	Days       *int          `json:"days"`
	Reason     *string       `json:"reason"`
	Status     string        `json:"status"`
	ReviewedBy *string       `json:"reviewedBy"`
	ReviewedAt *string       `json:"reviewedAt"`
	CreatedAt  time.Time     `json:"createdAt"`
	User       *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user,omitempty"`
}

// GetLeavesResponse represents the get leaves response
type GetLeavesResponse struct {
	Leaves []*LeaveResponse `json:"leaves"`
	Total  int              `json:"total"`
}

// CreatedLeave is the leave record as returned directly by POST /leaves (a raw row, not the joined list shape).
type CreatedLeave struct {
	ID        string  `json:"id"`
	MemberID  string  `json:"memberId"`
	TypeID    string  `json:"typeId"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Days      *int    `json:"days"`
	Reason    *string `json:"reason"`
	Status    string  `json:"status"`
}

// LeaveTypeBalance is one leave type's quota/usage/remaining for a year.
type LeaveTypeBalance struct {
	TypeID      string `json:"typeId"`
	Type        string `json:"type"`
	ResetPolicy string `json:"resetPolicy"`
	Year        int    `json:"year"`
	Quota       *int   `json:"quota"`
	Used        int    `json:"used"`
	Remaining   *int   `json:"remaining"`
}

// CreateLeaveResponse is returned by POST /leaves
type CreateLeaveResponse struct {
	Leave   *CreatedLeave     `json:"leave"`
	Balance *LeaveTypeBalance `json:"balance"`
}

// LeaveBalanceResponse is returned by GET /leaves/balance
type LeaveBalanceResponse struct {
	Year     int                 `json:"year"`
	Balances []*LeaveTypeBalance `json:"balances"`
}

// CreateLeave creates a new leave request
func (c *Client) CreateLeave(req *CreateLeaveRequest) (*CreateLeaveResponse, error) {
	var result CreateLeaveResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/leaves")

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
		Get("/api/v1/leaves")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetLeaveBalance retrieves the current user's leave balance (quota/used/remaining) per
// active leave type, for the given year. year of 0 uses the server's default (current year).
func (c *Client) GetLeaveBalance(year int) (*LeaveBalanceResponse, error) {
	req := c.http.R()
	if year > 0 {
		req.SetQueryParam("year", fmt.Sprintf("%d", year))
	}

	var result LeaveBalanceResponse
	resp, err := req.
		SetResult(&result).
		Get("/api/v1/leaves/balance")

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
		Delete(fmt.Sprintf("/api/v1/leaves/%s", leaveID))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return parseError(resp)
	}

	return nil
}

package api

import (
	"fmt"
	"time"
)

// CheckInRequest represents the check-in request
type CheckInRequest struct {
	Status string `json:"status,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// CheckOutRequest represents the check-out request
type CheckOutRequest struct {
	Notes string `json:"notes,omitempty"`
}

// AttendanceResponse represents an attendance record in API response
type AttendanceResponse struct {
	ID         string    `json:"id"`
	MemberID   string    `json:"memberId"`
	Date       string    `json:"date"`
	Status     string    `json:"status"`
	Notes      *string   `json:"notes"`
	CheckinAt  *string   `json:"checkinAt"`
	CheckoutAt *string   `json:"checkoutAt"`
	CreatedAt  time.Time `json:"createdAt"`
	User       *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user,omitempty"`
}

// GetMyAttendanceResponse represents the get my attendance response
type GetMyAttendanceResponse struct {
	Attendance []*AttendanceResponse `json:"attendance"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
}

// GetMyAttendance retrieves the current user's attendance history.
// period, if non-empty, must be one of "today", "week", or "month".
func (c *Client) GetMyAttendance(period string, limit, offset int) (*GetMyAttendanceResponse, error) {
	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}
	if period != "" {
		params["period"] = period
	}

	var result GetMyAttendanceResponse
	resp, err := c.http.R().
		SetQueryParams(params).
		SetResult(&result).
		Get("/api/v1/attendances/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// CheckIn checks in for the day
func (c *Client) CheckIn(req *CheckInRequest) (*AttendanceResponse, error) {
	var result AttendanceResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/attendances/checkin")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// CheckOut checks out for the day
func (c *Client) CheckOut(req *CheckOutRequest) (*AttendanceResponse, error) {
	var result AttendanceResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/attendances/checkout")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

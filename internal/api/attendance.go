package api

import (
	"fmt"
	"time"
)

// AttendanceRequest represents the attendance submission request
type AttendanceRequest struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Notes  string `json:"notes,omitempty"`
}

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
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Date         string    `json:"date"`
	Status       string    `json:"status"`
	Notes        *string   `json:"notes"`
	CheckinTime  *string   `json:"checkinTime"`
	CheckoutTime *string   `json:"checkoutTime"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	User         *struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user,omitempty"`
}

// NotMarkedUser represents a member who hasn't marked attendance yet
type NotMarkedUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetTodayAttendanceResponse represents the get today attendance response
type GetTodayAttendanceResponse struct {
	Attendance   []*AttendanceResponse `json:"attendance"`
	Date         string                `json:"date"`
	Marked       int                   `json:"marked"`
	TotalMembers int                   `json:"totalMembers"`
	StatusCounts map[string]int        `json:"statusCounts"`
	NotMarked    []*NotMarkedUser      `json:"notMarked"`
}

// GetMyAttendanceResponse represents the get my attendance response
type GetMyAttendanceResponse struct {
	Attendance []*AttendanceResponse `json:"attendance"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
}

// MarkAttendance marks attendance for a date
func (c *Client) MarkAttendance(req *AttendanceRequest) (*AttendanceResponse, error) {
	var result AttendanceResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/attendance")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetTodayAttendance retrieves all team attendance for today
func (c *Client) GetTodayAttendance() (*GetTodayAttendanceResponse, error) {
	var result GetTodayAttendanceResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/attendance/today")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetMyAttendance retrieves the current user's attendance history
func (c *Client) GetMyAttendance(limit, offset int) (*GetMyAttendanceResponse, error) {
	var result GetMyAttendanceResponse
	resp, err := c.http.R().
		SetQueryParams(map[string]string{
			"limit":  fmt.Sprintf("%d", limit),
			"offset": fmt.Sprintf("%d", offset),
		}).
		SetResult(&result).
		Get("/api/attendance/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// GetAttendanceByDate retrieves team attendance for a specific date
func (c *Client) GetAttendanceByDate(date string) (*GetTodayAttendanceResponse, error) {
	var result GetTodayAttendanceResponse
	resp, err := c.http.R().
		SetQueryParam("date", date).
		SetResult(&result).
		Get("/api/attendance")

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
		Post("/api/attendance/checkin")

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
		Post("/api/attendance/checkout")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

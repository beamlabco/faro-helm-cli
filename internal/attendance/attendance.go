package attendance

import (
	"fmt"
	"strings"

	"github.com/beamlabco/faro-helm-cli/internal/api"
)

// Service handles attendance operations
type Service struct {
	client *api.Client
}

// NewService creates a new attendance service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetToday retrieves the current user's attendance for today, if marked.
// Returns nil if attendance hasn't been marked yet today.
func (s *Service) GetToday() (*api.AttendanceResponse, error) {
	resp, err := s.client.GetMyAttendance("today", 1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's attendance: %w", err)
	}
	if len(resp.Attendance) == 0 {
		return nil, nil
	}

	return resp.Attendance[0], nil
}

// GetMy retrieves the current user's attendance history
func (s *Service) GetMy(limit, offset int) ([]*api.AttendanceResponse, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := s.client.GetMyAttendance("", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get my attendance: %w", err)
	}

	return resp.Attendance, resp.Total, nil
}

// CheckIn checks in for the day
func (s *Service) CheckIn(status, notes string) (*api.AttendanceResponse, error) {
	// Default to present if no status provided
	if status == "" {
		status = "present"
	}

	// Validate status (only present or remote allowed for checkin)
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "present" && status != "remote" {
		return nil, fmt.Errorf("check-in status must be 'present' or 'remote'")
	}

	req := &api.CheckInRequest{
		Status: status,
		Notes:  strings.TrimSpace(notes),
	}

	resp, err := s.client.CheckIn(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check in: %w", err)
	}

	return resp, nil
}

// CheckOut checks out for the day
func (s *Service) CheckOut(notes string) (*api.AttendanceResponse, error) {
	req := &api.CheckOutRequest{
		Notes: strings.TrimSpace(notes),
	}

	resp, err := s.client.CheckOut(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check out: %w", err)
	}

	return resp, nil
}

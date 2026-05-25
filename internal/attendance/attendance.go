package attendance

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm/internal/api"
)

// Valid attendance statuses
var ValidStatuses = []string{"present", "absent", "half-day", "remote"}

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

// Mark marks attendance for a date
func (s *Service) Mark(date, status, notes string) (*api.AttendanceResponse, error) {
	// Validate date format
	if date == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
		}
	}

	// Validate status
	status = strings.ToLower(strings.TrimSpace(status))
	if !isValidStatus(status) {
		return nil, fmt.Errorf("invalid status, must be one of: %s", strings.Join(ValidStatuses, ", "))
	}

	// Trim notes
	notes = strings.TrimSpace(notes)

	req := &api.AttendanceRequest{
		Date:   date,
		Status: status,
		Notes:  notes,
	}

	resp, err := s.client.MarkAttendance(req)
	if err != nil {
		return nil, fmt.Errorf("failed to mark attendance: %w", err)
	}

	return resp, nil
}

// GetToday retrieves all team attendance for today
func (s *Service) GetToday() (*api.GetTodayAttendanceResponse, error) {
	resp, err := s.client.GetTodayAttendance()
	if err != nil {
		return nil, fmt.Errorf("failed to get today's attendance: %w", err)
	}

	return resp, nil
}

// GetMy retrieves the current user's attendance history
func (s *Service) GetMy(limit, offset int) ([]*api.AttendanceResponse, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := s.client.GetMyAttendance(limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get my attendance: %w", err)
	}

	return resp.Attendance, resp.Total, nil
}

// GetByDate retrieves team attendance for a specific date
func (s *Service) GetByDate(date string) (*api.GetTodayAttendanceResponse, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	resp, err := s.client.GetAttendanceByDate(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance for date %s: %w", date, err)
	}

	return resp, nil
}

func isValidStatus(status string) bool {
	for _, s := range ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
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

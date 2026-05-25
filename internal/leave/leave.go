package leave

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm/internal/api"
)

// Valid leave types
var ValidTypes = []string{"sick", "casual", "paid", "unpaid", "wfh"}

// Valid leave statuses
var ValidStatuses = []string{"pending", "approved", "rejected"}

// Service handles leave operations
type Service struct {
	client *api.Client
}

// NewService creates a new leave service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// Create creates a new leave request
func (s *Service) Create(leaveType, startDate, endDate, reason string) (*api.LeaveResponse, error) {
	// Validate type
	leaveType = strings.ToLower(strings.TrimSpace(leaveType))
	if !isValidType(leaveType) {
		return nil, fmt.Errorf("invalid leave type, must be one of: %s", strings.Join(ValidTypes, ", "))
	}

	// Validate dates
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, fmt.Errorf("invalid start date format, use YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return nil, fmt.Errorf("invalid end date format, use YYYY-MM-DD")
	}
	if endDate < startDate {
		return nil, fmt.Errorf("end date must be on or after start date")
	}

	reason = strings.TrimSpace(reason)

	req := &api.CreateLeaveRequest{
		Type:      leaveType,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
	}

	resp, err := s.client.CreateLeave(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave request: %w", err)
	}

	return resp, nil
}

// GetAll retrieves leaves with optional filters
func (s *Service) GetAll(status string, userID string, limit, offset int) ([]*api.LeaveResponse, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := s.client.GetLeaves(status, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leaves: %w", err)
	}

	return resp.Leaves, resp.Count, nil
}

// UpdateStatus approves or rejects a leave request
func (s *Service) UpdateStatus(leaveID string, status string) (*api.LeaveResponse, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "approved" && status != "rejected" {
		return nil, fmt.Errorf("status must be 'approved' or 'rejected'")
	}

	resp, err := s.client.UpdateLeaveStatus(leaveID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update leave status: %w", err)
	}

	return resp, nil
}

// Cancel cancels a pending leave request
func (s *Service) Cancel(leaveID string) error {
	if err := s.client.CancelLeave(leaveID); err != nil {
		return fmt.Errorf("failed to cancel leave: %w", err)
	}

	return nil
}

func isValidType(t string) bool {
	for _, v := range ValidTypes {
		if v == t {
			return true
		}
	}
	return false
}

package leave

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm-cli/internal/api"
)

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

// GetBalance retrieves the current user's leave quota/used/remaining per active leave
// type, for the given year. year of 0 uses the server's default (current year).
func (s *Service) GetBalance(year int) ([]*api.LeaveTypeBalance, int, error) {
	resp, err := s.client.GetLeaveBalance(year)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leave balance: %w", err)
	}

	return resp.Balances, resp.Year, nil
}

// Create creates a new leave request for the given leave type ID
func (s *Service) Create(typeID, startDate, endDate, reason string) (*api.CreatedLeave, error) {
	typeID = strings.TrimSpace(typeID)
	if typeID == "" {
		return nil, fmt.Errorf("a leave type must be selected")
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
		TypeID:    typeID,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
	}

	resp, err := s.client.CreateLeave(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create leave request: %w", err)
	}

	return resp.Leave, nil
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

	return resp.Leaves, resp.Total, nil
}

// Cancel cancels a pending leave request
func (s *Service) Cancel(leaveID string) error {
	if err := s.client.CancelLeave(leaveID); err != nil {
		return fmt.Errorf("failed to cancel leave: %w", err)
	}

	return nil
}

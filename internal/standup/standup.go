package standup

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm/internal/api"
)

// Service handles standup operations
type Service struct {
	client *api.Client
}

// NewService creates a new standup service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// Submit submits or updates a standup
func (s *Service) Submit(projectID string, date, yesterday, today, blockers string) (*api.StandupResponse, error) {
	// Validate date format
	if date == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
		}
	}

	// Trim whitespace
	yesterday = strings.TrimSpace(yesterday)
	today = strings.TrimSpace(today)
	blockers = strings.TrimSpace(blockers)

	// At least one field must be filled
	if yesterday == "" && today == "" && blockers == "" {
		return nil, fmt.Errorf("at least one field (yesterday, today, or blockers) must be filled")
	}

	req := &api.StandupRequest{
		Date:      date,
		ProjectID: projectID,
		Yesterday: yesterday,
		Today:     today,
		Blockers:  blockers,
	}

	resp, err := s.client.SubmitStandup(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit standup: %w", err)
	}

	return resp.Standup, nil
}

// GetToday retrieves all team standups for today
func (s *Service) GetToday() ([]*api.StandupResponse, string, error) {
	resp, err := s.client.GetTodayStandups()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get today's standups: %w", err)
	}

	return resp.Standups, resp.Date, nil
}

// GetMy retrieves the current user's standup history
func (s *Service) GetMy(limit, offset int) ([]*api.StandupResponse, *api.Pagination, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := s.client.GetMyStandups(limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get my standups: %w", err)
	}

	return resp.Standups, resp.Pagination, nil
}

// GetByDate retrieves team standups for a specific date
func (s *Service) GetByDate(date string) ([]*api.StandupResponse, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	resp, err := s.client.GetStandupsByDate(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get standups for date %s: %w", date, err)
	}

	return resp.Standups, nil
}

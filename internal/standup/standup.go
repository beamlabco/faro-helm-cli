package standup

import (
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm-cli/internal/api"
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
func (s *Service) Submit(teamID string, date, yesterday, today, blockers string) (*api.StandupResponse, error) {
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
		TeamID:    teamID,
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

// GetToday retrieves the current user's standups submitted today (across teams).
func (s *Service) GetToday() ([]*api.StandupResponse, error) {
	resp, err := s.client.GetMyStandups("today", 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's standups: %w", err)
	}

	return resp.Standups, nil
}

// GetMy retrieves the current user's standup history. Returns standups, total count, error.
func (s *Service) GetMy(limit, offset int) ([]*api.StandupResponse, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	resp, err := s.client.GetMyStandups("", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get my standups: %w", err)
	}

	return resp.Standups, resp.Total, nil
}

package team

import (
	"fmt"

	"github.com/beamlabco/faro-helm-cli/internal/api"
)

// Service handles team operations
type Service struct {
	client *api.Client
}

// NewService creates a new team service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetMyTeams retrieves the current user's teams
func (s *Service) GetMyTeams() ([]*api.TeamResponse, error) {
	teams, err := s.client.GetMyTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teams, nil
}

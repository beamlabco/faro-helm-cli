package project

import (
	"fmt"

	"github.com/beamlabco/faro-helm-cli/internal/api"
)

// Service handles project operations
type Service struct {
	client *api.Client
}

// NewService creates a new project service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetMyProjects retrieves the current user's projects
func (s *Service) GetMyProjects() ([]*api.ProjectResponse, error) {
	projects, err := s.client.GetMyProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}

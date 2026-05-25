package organization

import (
	"fmt"

	"github.com/beamlabco/faro-helm/internal/api"
)

// Service handles organization operations
type Service struct {
	client *api.Client
}

// NewService creates a new organization service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetSettings retrieves organization settings
func (s *Service) GetSettings() (*api.OrgSettingsResponse, error) {
	settings, err := s.client.GetOrgSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get organization settings: %w", err)
	}

	return settings, nil
}

// UpdateSettings updates organization settings
func (s *Service) UpdateSettings(req *api.UpdateOrgSettingsRequest) (*api.OrgSettingsResponse, error) {
	settings, err := s.client.UpdateOrgSettings(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization settings: %w", err)
	}

	return settings, nil
}

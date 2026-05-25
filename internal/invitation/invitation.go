package invitation

import (
	"fmt"
	"strings"

	"github.com/beamlabco/faro-helm/internal/api"
)

// Service handles invitation operations
type Service struct {
	client *api.Client
}

// NewService creates a new invitation service
func NewService(client *api.Client) *Service {
	return &Service{
		client: client,
	}
}

// Create creates an invitation for the given email
func (s *Service) Create(email string) (*api.InvitationResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}

	req := &api.CreateInvitationRequest{
		Email: email,
	}

	resp, err := s.client.CreateInvitation(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return resp.Invitation, nil
}

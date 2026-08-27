package user

import (
	"fmt"

	"github.com/beamlabco/faro-helm-cli/internal/api"
)

// Service handles user operations
type Service struct {
	client     *api.Client
	authClient *api.AuthClient
	token      func() string
}

// NewService creates a new user service
func NewService(client *api.Client, authClient *api.AuthClient, token func() string) *Service {
	return &Service{client: client, authClient: authClient, token: token}
}

// GetMembers retrieves all members of the organization
func (s *Service) GetMembers() ([]*api.MemberResponse, error) {
	resp, err := s.client.GetMembers()
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	return resp.Users, nil
}

// ChangePassword changes the authenticated user's password via auth-api.
func (s *Service) ChangePassword(currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}

	if err := s.authClient.ChangePassword(s.token(), currentPassword, newPassword); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

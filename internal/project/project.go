package project

import (
	"fmt"

	"github.com/beamlabco/faro-helm/internal/api"
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

// GetAll retrieves all projects in the organization
func (s *Service) GetAll() ([]*api.ProjectResponse, error) {
	projects, err := s.client.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	return projects, nil
}

// Get retrieves a single project
func (s *Service) Get(projectID string) (*api.ProjectResponse, error) {
	project, err := s.client.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// Create creates a new project
func (s *Service) Create(name string) (*api.ProjectResponse, error) {
	project, err := s.client.CreateProject(&api.CreateProjectRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return project, nil
}

// Update updates project settings
func (s *Service) Update(projectID string, req *api.UpdateProjectRequest) (*api.ProjectResponse, error) {
	project, err := s.client.UpdateProject(projectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	return project, nil
}

// Delete deletes a project
func (s *Service) Delete(projectID string) error {
	if err := s.client.DeleteProject(projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// GetMembers retrieves project members
func (s *Service) GetMembers(projectID string) ([]*api.ProjectMemberResponse, error) {
	members, err := s.client.GetProjectMembers(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}
	return members, nil
}

// AddMembers adds members to a project
func (s *Service) AddMembers(projectID string, userIDs []string) error {
	if err := s.client.AddProjectMembers(projectID, userIDs); err != nil {
		return fmt.Errorf("failed to add members: %w", err)
	}
	return nil
}

// RemoveMember removes a member from a project
func (s *Service) RemoveMember(projectID, userID string) error {
	if err := s.client.RemoveProjectMember(projectID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

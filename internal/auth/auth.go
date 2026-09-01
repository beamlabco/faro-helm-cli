package auth

import (
	"fmt"
	"strings"

	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/config"
)

// Service handles authentication operations
type Service struct {
	client *api.Client
	config *config.Config
}

// NewService creates a new authentication service
func NewService(client *api.Client, cfg *config.Config) *Service {
	return &Service{
		client: client,
		config: cfg,
	}
}

// Register creates a new account and workspace.
// No token is returned — the user must verify their email then run /login.
func (s *Service) Register(email, password, name, organizationName string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateOrganizationName(organizationName); err != nil {
		return err
	}

	return s.client.Register(&api.AuthRegisterRequest{
		Email:            email,
		Password:         password,
		Name:             name,
		OrganizationName: organizationName,
	})
}

// Join accepts a workspace invitation and logs the user in immediately.
func (s *Service) Join(email, password, name, token string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	if err := validateName(name); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("invitation token is required")
	}

	resp, err := s.client.AcceptInvitation(&api.AcceptInvitationRequest{
		Email:    email,
		Password: password,
		Name:     name,
		Token:    token,
	})
	if err != nil {
		return fmt.Errorf("join failed: %w", err)
	}

	user := &config.User{
		ID:        resp.Member.ID,
		AccountID: resp.Account.ID,
		Email:     resp.Account.Email,
		Name:      resp.Account.Name,
		Role:      resp.Member.Role,
	}
	org := &config.Organization{
		ID:     resp.Workspace.ID,
		Name:   resp.Workspace.Name,
		Status: resp.Workspace.Status,
	}

	s.config.SetAuthData(resp.AccessToken, resp.RefreshToken, user, org)
	if err := config.Save(s.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.client.SetToken(resp.AccessToken)
	return nil
}

// Login authenticates and stores the session.
func (s *Service) Login(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	resp, err := s.client.Login(&api.AuthLoginRequest{Email: email, Password: password})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	me, err := s.client.GetMe(resp.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch account info: %w", err)
	}
	if len(me.Workspaces) == 0 {
		return fmt.Errorf("account has no workspaces — register or join an organisation first")
	}

	ws := me.Workspaces[0]
	user := &config.User{
		ID:        ws.MemberID,
		AccountID: resp.Account.ID,
		Email:     resp.Account.Email,
		Name:      resp.Account.Name,
		Role:      ws.Role,
	}
	org := &config.Organization{
		ID:     ws.WorkspaceID,
		Name:   ws.Name,
		Status: "active",
	}

	s.config.SetAuthData(resp.AccessToken, resp.RefreshToken, user, org)
	if err := config.Save(s.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.client.SetToken(resp.AccessToken)
	return nil
}

// Logout clears the authentication data
func (s *Service) Logout() error {
	s.config.Clear()
	if err := config.Save(s.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.client.SetToken("")
	return nil
}

// IsAuthenticated checks if the user is authenticated
func (s *Service) IsAuthenticated() bool {
	return s.config.IsAuthenticated()
}

// GetUser returns the authenticated user
func (s *Service) GetUser() *config.User {
	return s.config.User
}

// GetOrganization returns the user's organization
func (s *Service) GetOrganization() *config.Organization {
	return s.config.Organization
}

// DeviceFlowInfo holds the display info returned after initiating device flow.
type DeviceFlowInfo struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       int
	Interval        int
}

// StartDeviceFlow initiates the device authorization flow and returns display info.
func (s *Service) StartDeviceFlow() (*DeviceFlowInfo, error) {
	resp, err := s.client.InitiateDeviceFlow()
	if err != nil {
		return nil, fmt.Errorf("failed to start device flow: %w", err)
	}
	return &DeviceFlowInfo{
		DeviceCode:      resp.DeviceCode,
		UserCode:        resp.UserCode,
		VerificationURI: resp.VerificationURI,
		ExpiresIn:       resp.ExpiresIn,
		Interval:        resp.Interval,
	}, nil
}

// PollDeviceToken polls once for a completed device authorization.
func (s *Service) PollDeviceToken(deviceCode string) (*api.DeviceTokenPair, error) {
	return s.client.PollDeviceToken(deviceCode)
}

// CompleteDeviceLogin saves auth data after a successful device flow poll.
func (s *Service) CompleteDeviceLogin(tokenPair *api.DeviceTokenPair) error {
	me, err := s.client.GetMe(tokenPair.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch account info: %w", err)
	}

	if len(me.Workspaces) == 0 {
		return fmt.Errorf("account has no workspaces — register or join an organisation first")
	}

	ws := me.Workspaces[0]
	user := &config.User{
		ID:        ws.MemberID,
		AccountID: me.Account.ID,
		Email:     me.Account.Email,
		Name:      me.Account.Name,
		Role:      ws.Role,
	}
	org := &config.Organization{
		ID:     ws.WorkspaceID,
		Name:   ws.Name,
		Status: "active",
	}

	s.config.SetAuthData(tokenPair.AccessToken, tokenPair.RefreshToken, user, org)
	if err := config.Save(s.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	s.client.SetToken(tokenPair.AccessToken)
	return nil
}

// Validation functions

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	return nil
}

func validateOrganizationName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("organisation name must be at least 2 characters")
	}
	return nil
}

package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/config"
	"github.com/beamlabco/faro-helm-cli/internal/oauthflow"
)

// browserLoginTimeout bounds how long CompleteBrowserLogin waits for the
// browser redirect before giving up.
const browserLoginTimeout = 5 * time.Minute

// Service handles authentication operations
type Service struct {
	client     *api.Client
	authClient *api.AuthClient
	config     *config.Config
}

// NewService creates a new authentication service
func NewService(client *api.Client, authClient *api.AuthClient, cfg *config.Config) *Service {
	return &Service{
		client:     client,
		authClient: authClient,
		config:     cfg,
	}
}

// Register creates a new account and workspace on auth-api.
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

	return s.authClient.Register(&api.AuthRegisterRequest{
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

	resp, err := s.authClient.AcceptInvitation(&api.AcceptInvitationRequest{
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

// BrowserLogin holds the in-progress state of an Authorization Code + PKCE
// login: the URL to open (or show the user, if auto-open fails) and the
// local callback server + verifier needed to complete it.
type BrowserLogin struct {
	AuthorizeURL string

	server   *oauthflow.CallbackServer
	verifier string
	state    string
}

// loginScope is requested from the auth server; it's intersected server-side
// against faro-helm-cli's registered allowed_scopes, so requesting more than
// the client is allowed is harmless.
const loginScope = "helm:read helm:write profile"

// BeginBrowserLogin starts the Authorization Code + PKCE flow: generates
// PKCE + state, binds the local loopback callback server, and builds the
// /oauth/authorize URL. It does not open a browser or block — callers
// typically want to render the URL immediately, then call OpenBrowser and
// CompleteBrowserLogin.
func (s *Service) BeginBrowserLogin() (*BrowserLogin, error) {
	pkce, err := oauthflow.GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}
	state, err := oauthflow.GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	server, err := oauthflow.StartCallbackServer()
	if err != nil {
		return nil, fmt.Errorf("failed to start local callback server: %w", err)
	}

	authorizeURL := oauthflow.BuildAuthorizeURL(oauthflow.AuthorizeURLParams{
		AuthBaseURL:   s.authClient.BaseURL(),
		ClientID:      oauthflow.ClientID,
		RedirectURI:   server.RedirectURI(),
		Scope:         loginScope,
		State:         state,
		CodeChallenge: pkce.Challenge,
	})

	return &BrowserLogin{
		AuthorizeURL: authorizeURL,
		server:       server,
		verifier:     pkce.Verifier,
		state:        state,
	}, nil
}

// CompleteBrowserLogin blocks until the browser redirect arrives (or
// browserLoginTimeout elapses / ctx is cancelled), exchanges the resulting
// code for tokens, resolves the account's workspace, and persists the
// session — mirroring what Login/CompleteDeviceLogin used to do. Always
// closes the callback server before returning.
func (s *Service) CompleteBrowserLogin(ctx context.Context, bl *BrowserLogin) error {
	defer bl.server.Close()

	ctx, cancel := context.WithTimeout(ctx, browserLoginTimeout)
	defer cancel()

	code, err := bl.server.Await(ctx, bl.state)
	if err != nil {
		return fmt.Errorf("sign-in failed: %w", err)
	}

	tokens, err := s.authClient.ExchangeAuthorizationCode(code, bl.server.RedirectURI(), bl.verifier)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	me, err := s.authClient.GetMe(tokens.AccessToken)
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

	s.config.SetAuthData(tokens.AccessToken, tokens.RefreshToken, user, org)
	if err := config.Save(s.config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	s.client.SetToken(tokens.AccessToken)
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

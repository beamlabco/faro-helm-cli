package api

import (
	"github.com/beamlabco/faro-helm/internal/config"
)

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	OrganizationName string `json:"organizationName,omitempty"`
	Token            string `json:"token,omitempty"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token        string                  `json:"token"`
	User         *UserResponse           `json:"user"`
	Organization *OrganizationResponse   `json:"organization"`
}

// UserResponse represents the user in API response
type UserResponse struct {
	ID             string `json:"id"`
	AccountID      string `json:"accountId"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
}

// OrganizationResponse represents the organization in API response
type OrganizationResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Register creates a new user and organization
func (c *Client) Register(req *RegisterRequest) (*AuthResponse, error) {
	var result AuthResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/auth/register")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// Login authenticates a user and returns a JWT token
func (c *Client) Login(req *LoginRequest) (*AuthResponse, error) {
	var result AuthResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/auth/login")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// ToConfigUser converts UserResponse to config.User
func (u *UserResponse) ToConfigUser() *config.User {
	return &config.User{
		ID:        u.ID,
		AccountID: u.AccountID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
	}
}

// ToConfigOrganization converts OrganizationResponse to config.Organization
func (o *OrganizationResponse) ToConfigOrganization() *config.Organization {
	return &config.Organization{
		ID:     o.ID,
		Name:   o.Name,
		Status: o.Status,
	}
}

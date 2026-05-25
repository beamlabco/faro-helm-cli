package api

import "time"

// CreateInvitationRequest represents the create invitation request
type CreateInvitationRequest struct {
	Email string `json:"email"`
}

// InvitationResponse represents an invitation in API response
type InvitationResponse struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	OrganizationID int       `json:"organizationId"`
	InvitedBy      int       `json:"invitedBy"`
	Status         string    `json:"status"`
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expiresAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateInvitationResponse wraps the invitation response
type CreateInvitationResponse struct {
	Invitation *InvitationResponse `json:"invitation"`
}

// CreateInvitation creates an invitation to join the organization
func (c *Client) CreateInvitation(req *CreateInvitationRequest) (*CreateInvitationResponse, error) {
	var result CreateInvitationResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/invitations")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

package api

// AcceptInvitationRequest is the body for POST /auth/invitations/accept.
type AcceptInvitationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

// AcceptInvitationResponse is returned by POST /auth/invitations/accept.
type AcceptInvitationResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Account      struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"account"`
	Workspace struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"workspace"`
	Member struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	} `json:"member"`
}

// AcceptInvitation joins a workspace via an invitation token.
func (c *AuthClient) AcceptInvitation(req *AcceptInvitationRequest) (*AcceptInvitationResponse, error) {
	var result AcceptInvitationResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/auth/invitations/accept")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, parseError(resp)
	}
	return &result, nil
}

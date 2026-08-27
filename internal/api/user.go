package api

// GetMembersResponse represents the get members response
type GetMembersResponse struct {
	Users []*MemberResponse `json:"users"`
}

// MemberResponse represents a team member in API response
type MemberResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// GetMembers retrieves all members of the current organization
func (c *Client) GetMembers() (*GetMembersResponse, error) {
	var result GetMembersResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/v1/users")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

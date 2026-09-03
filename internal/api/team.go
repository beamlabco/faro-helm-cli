package api

import "time"

// TeamResponse represents a team in API response
type TeamResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Timezone    *string   `json:"timezone"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type teamsListResponse struct {
	Teams []*TeamResponse `json:"teams"`
}

// GetMyTeams retrieves the current user's teams
func (c *Client) GetMyTeams() ([]*TeamResponse, error) {
	var result teamsListResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/v1/teams/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return result.Teams, nil
}

package api

import "time"

// ProjectResponse represents a project in API response
type ProjectResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Timezone    *string   `json:"timezone"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type projectsListResponse struct {
	Projects []*ProjectResponse `json:"projects"`
}

// GetMyProjects retrieves the current user's projects
func (c *Client) GetMyProjects() ([]*ProjectResponse, error) {
	var result projectsListResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/v1/projects/my")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return result.Projects, nil
}

package api

// OrgSettingsResponse represents the organization settings
type OrgSettingsResponse struct {
	Name               string  `json:"name"`
	Timezone           *string `json:"timezone"`
	OfficeStartTime    *string `json:"officeStartTime"`
	GracePeriodMinutes *int    `json:"gracePeriodMinutes"`
}

// UpdateOrgSettingsRequest represents the update settings request
type UpdateOrgSettingsRequest struct {
	Name               *string `json:"name,omitempty"`
	Timezone           *string `json:"timezone,omitempty"`
	OfficeStartTime    *string `json:"officeStartTime,omitempty"`
	GracePeriodMinutes *int    `json:"gracePeriodMinutes,omitempty"`
}

// GetOrgSettings retrieves organization settings
func (c *Client) GetOrgSettings() (*OrgSettingsResponse, error) {
	var result OrgSettingsResponse
	resp, err := c.http.R().
		SetResult(&result).
		Get("/api/organizations/settings")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

// UpdateOrgSettings updates organization settings
func (c *Client) UpdateOrgSettings(req *UpdateOrgSettingsRequest) (*OrgSettingsResponse, error) {
	var result OrgSettingsResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Patch("/api/organizations/settings")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, parseError(resp)
	}

	return &result, nil
}

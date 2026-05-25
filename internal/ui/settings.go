package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/organization"
)

const (
	settingsOrgName = iota
	settingsTimezone
	settingsOfficeStartTime
	settingsGracePeriodMinutes
)

// SettingsModel represents the organization settings form
type SettingsModel struct {
	orgService   *organization.Service
	inputs       []textinput.Model
	focusIndex   int
	errorMsg     string
	successMsg   string
	loading      bool
	shouldGoBack bool
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(orgService *organization.Service) SettingsModel {
	orgName := textinput.New()
	orgName.Placeholder = "Organization name"
	orgName.CharLimit = 255
	orgName.Width = 40

	timezone := textinput.New()
	timezone.Placeholder = "America/New_York"
	timezone.CharLimit = 50
	timezone.Width = 30

	officeStartTime := textinput.New()
	officeStartTime.Placeholder = "09:00  (leave blank to disable)"
	officeStartTime.CharLimit = 5
	officeStartTime.Width = 30

	gracePeriodMinutes := textinput.New()
	gracePeriodMinutes.Placeholder = "10"
	gracePeriodMinutes.CharLimit = 2
	gracePeriodMinutes.Width = 10

	return SettingsModel{
		orgService: orgService,
		inputs: []textinput.Model{
			orgName,
			timezone,
			officeStartTime,
			gracePeriodMinutes,
		},
		focusIndex: settingsOrgName,
		loading:    true,
	}
}

// Init loads current settings
func (m SettingsModel) Init() tea.Cmd {
	return m.fetchSettings()
}

// Update handles messages
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "tab", "down":
			if !m.loading {
				m.inputs[m.focusIndex].Blur()
				m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
				m.inputs[m.focusIndex].Focus()
			}
			return m, nil

		case "shift+tab", "up":
			if !m.loading {
				m.inputs[m.focusIndex].Blur()
				m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
				m.inputs[m.focusIndex].Focus()
			}
			return m, nil

		case "ctrl+s", "enter":
			if m.loading {
				return m, nil
			}
			return m, m.handleSubmit()
		}

	case settingsLoadedMsg:
		m.loading = false
		m.populateFromSettings(msg.settings)
		m.inputs[settingsOrgName].Focus()
		return m, nil

	case settingsUpdatedMsg:
		m.loading = false
		m.successMsg = "Settings updated successfully!"
		m.errorMsg = ""
		return m, nil

	case settingsErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update focused text input
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *SettingsModel) populateFromSettings(s *api.OrgSettingsResponse) {
	m.inputs[settingsOrgName].SetValue(s.Name)
	if s.Timezone != nil {
		m.inputs[settingsTimezone].SetValue(*s.Timezone)
	}
	if s.OfficeStartTime != nil {
		m.inputs[settingsOfficeStartTime].SetValue(*s.OfficeStartTime)
	}
	if s.GracePeriodMinutes != nil {
		m.inputs[settingsGracePeriodMinutes].SetValue(strconv.Itoa(*s.GracePeriodMinutes))
	}
}

// View renders the settings form
func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Organization Settings"))
	b.WriteString("\n\n")

	if m.loading && m.successMsg == "" && m.errorMsg == "" {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading settings..."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	// Organization name
	b.WriteString(labelStyle.Render("Organization name:"))
	b.WriteString("\n")
	if m.focusIndex == settingsOrgName {
		b.WriteString(focusedInputStyle.Render(m.inputs[settingsOrgName].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[settingsOrgName].View()))
	}
	b.WriteString("\n\n")

	// Timezone
	b.WriteString(labelStyle.Render("Timezone (IANA):"))
	b.WriteString("\n")
	if m.focusIndex == settingsTimezone {
		b.WriteString(focusedInputStyle.Render(m.inputs[settingsTimezone].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[settingsTimezone].View()))
	}
	b.WriteString("\n\n")

	// Office hours section
	b.WriteString(labelStyle.Render("Office hours"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Used to detect late check-ins in the weekly attendance summary."))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Office start time (HH:MM):"))
	b.WriteString("\n")
	if m.focusIndex == settingsOfficeStartTime {
		b.WriteString(focusedInputStyle.Render(m.inputs[settingsOfficeStartTime].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[settingsOfficeStartTime].View()))
	}
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Grace period (minutes, 0–60):"))
	b.WriteString("\n")
	if m.focusIndex == settingsGracePeriodMinutes {
		b.WriteString(focusedInputStyle.Render(m.inputs[settingsGracePeriodMinutes].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[settingsGracePeriodMinutes].View()))
	}
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Saving settings..."))
		b.WriteString("\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("[↑↓/Tab] Navigate  [Enter/Ctrl+S] Save  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *SettingsModel) fetchSettings() tea.Cmd {
	return func() tea.Msg {
		settings, err := m.orgService.GetSettings()
		if err != nil {
			return settingsErrorMsg(err.Error())
		}
		return settingsLoadedMsg{settings: settings}
	}
}

func (m *SettingsModel) handleSubmit() tea.Cmd {
	orgName := strings.TrimSpace(m.inputs[settingsOrgName].Value())
	timezone := strings.TrimSpace(m.inputs[settingsTimezone].Value())
	officeStartTime := strings.TrimSpace(m.inputs[settingsOfficeStartTime].Value())
	gracePeriodStr := strings.TrimSpace(m.inputs[settingsGracePeriodMinutes].Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		req := &api.UpdateOrgSettingsRequest{}

		if orgName != "" {
			req.Name = &orgName
		}
		if timezone != "" {
			req.Timezone = &timezone
		}
		if officeStartTime != "" {
			req.OfficeStartTime = &officeStartTime
		}
		if gracePeriodStr != "" {
			minutes, err := strconv.Atoi(gracePeriodStr)
			if err != nil || minutes < 0 || minutes > 60 {
				return settingsErrorMsg("Grace period must be a number between 0 and 60")
			}
			req.GracePeriodMinutes = &minutes
		}

		_, err := m.orgService.UpdateSettings(req)
		if err != nil {
			return settingsErrorMsg(err.Error())
		}
		return settingsUpdatedMsg{}
	}
}

// Message types
type settingsLoadedMsg struct {
	settings *api.OrgSettingsResponse
}
type settingsUpdatedMsg struct{}
type settingsErrorMsg string

package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/project"
)

const (
	projSettingsName = iota
	projSettingsEnabled
	projSettingsSummaryTime
	projSettingsTimezone
	projSettingsDiscordURL
	projSettingsWeeklySummaryDay
	projSettingsWeeklySummaryTime
)

// ProjectSettingsModel represents the project settings form
type ProjectSettingsModel struct {
	projectService *project.Service
	projects       []*api.ProjectResponse
	selectedIdx    int
	inputs         []textinput.Model
	enabledToggle  bool
	focusIndex     int
	phase          string // "select" or "edit"
	errorMsg       string
	successMsg     string
	loading        bool
	shouldGoBack   bool
}

// NewProjectSettingsModel creates a new project settings model
func NewProjectSettingsModel(projectService *project.Service) ProjectSettingsModel {
	name := textinput.New()
	name.Placeholder = "Project name"
	name.CharLimit = 255
	name.Width = 40

	summaryTime := textinput.New()
	summaryTime.Placeholder = "09:00"
	summaryTime.CharLimit = 5
	summaryTime.Width = 20

	timezone := textinput.New()
	timezone.Placeholder = "America/New_York"
	timezone.CharLimit = 50
	timezone.Width = 30

	discordURL := textinput.New()
	discordURL.Placeholder = "https://discord.com/api/webhooks/..."
	discordURL.CharLimit = 500
	discordURL.Width = 55

	weeklySummaryDay := textinput.New()
	weeklySummaryDay.Placeholder = "5  (1=Mon … 7=Sun, blank to disable)"
	weeklySummaryDay.CharLimit = 1
	weeklySummaryDay.Width = 38

	weeklySummaryTime := textinput.New()
	weeklySummaryTime.Placeholder = "17:00  (defaults to daily summary time)"
	weeklySummaryTime.CharLimit = 5
	weeklySummaryTime.Width = 38

	return ProjectSettingsModel{
		projectService: projectService,
		inputs: []textinput.Model{
			name,            // inputIndex 0
			summaryTime,     // inputIndex 1
			timezone,        // inputIndex 2
			discordURL,      // inputIndex 3
			weeklySummaryDay,  // inputIndex 4
			weeklySummaryTime, // inputIndex 5
		},
		enabledToggle: true,
		focusIndex:    projSettingsName,
		phase:         "select",
		loading:       true,
	}
}

// Init loads projects
func (m ProjectSettingsModel) Init() tea.Cmd {
	return m.fetchProjects()
}

// Update handles messages
func (m ProjectSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.phase == "edit" {
				m.phase = "select"
				m.errorMsg = ""
				m.successMsg = ""
				return m, nil
			}
			m.shouldGoBack = true
			return m, nil

		case "up", "k":
			if m.phase == "select" && m.selectedIdx > 0 {
				m.selectedIdx--
			}
			if m.phase == "edit" {
				m.blurAll()
				m.focusIndex--
				if m.focusIndex < projSettingsName {
					m.focusIndex = projSettingsWeeklySummaryTime
				}
				m.focusCurrent()
			}
			return m, nil

		case "down", "j":
			if m.phase == "select" && m.selectedIdx < len(m.projects)-1 {
				m.selectedIdx++
			}
			if m.phase == "edit" {
				m.blurAll()
				m.focusIndex++
				if m.focusIndex > projSettingsWeeklySummaryTime {
					m.focusIndex = projSettingsName
				}
				m.focusCurrent()
			}
			return m, nil

		case "enter":
			if m.phase == "select" && len(m.projects) > 0 {
				m.phase = "edit"
				m.populateFromProject(m.projects[m.selectedIdx])
				m.inputs[inputIndex(projSettingsName)].Focus()
				return m, nil
			}
			if m.phase == "edit" && !m.loading {
				return m, m.handleSubmit()
			}

		case "tab":
			if m.phase == "edit" {
				m.blurAll()
				m.focusIndex++
				if m.focusIndex > projSettingsWeeklySummaryTime {
					m.focusIndex = projSettingsName
				}
				m.focusCurrent()
				return m, nil
			}

		case "shift+tab":
			if m.phase == "edit" {
				m.blurAll()
				m.focusIndex--
				if m.focusIndex < projSettingsName {
					m.focusIndex = projSettingsWeeklySummaryTime
				}
				m.focusCurrent()
				return m, nil
			}

		case " ":
			if m.phase == "edit" && m.focusIndex == projSettingsEnabled {
				m.enabledToggle = !m.enabledToggle
				return m, nil
			}

		case "ctrl+s":
			if m.phase == "edit" && !m.loading {
				return m, m.handleSubmit()
			}
		}

	case projectSettingsProjectsMsg:
		m.loading = false
		m.projects = msg.projects
		return m, nil

	case projectSettingsUpdatedMsg:
		m.loading = false
		m.successMsg = "Project settings updated!"
		m.errorMsg = ""
		return m, nil

	case projectSettingsErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update focused text input in edit phase
	if m.phase == "edit" {
		idx := inputIndex(m.focusIndex)
		if idx >= 0 {
			var cmd tea.Cmd
			m.inputs[idx], cmd = m.inputs[idx].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// inputIndex maps a focus constant to the inputs array index.
// projSettingsEnabled is a toggle, not in the inputs array.
func inputIndex(focusIdx int) int {
	switch focusIdx {
	case projSettingsName:
		return 0
	case projSettingsSummaryTime:
		return 1
	case projSettingsTimezone:
		return 2
	case projSettingsDiscordURL:
		return 3
	case projSettingsWeeklySummaryDay:
		return 4
	case projSettingsWeeklySummaryTime:
		return 5
	default:
		return -1
	}
}

func (m *ProjectSettingsModel) blurAll() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func (m *ProjectSettingsModel) focusCurrent() {
	idx := inputIndex(m.focusIndex)
	if idx >= 0 {
		m.inputs[idx].Focus()
	}
}

func (m *ProjectSettingsModel) populateFromProject(p *api.ProjectResponse) {
	m.inputs[inputIndex(projSettingsName)].SetValue(p.Name)
	if p.SummaryTime != nil {
		m.inputs[inputIndex(projSettingsSummaryTime)].SetValue(*p.SummaryTime)
	}
	if p.Timezone != nil {
		m.inputs[inputIndex(projSettingsTimezone)].SetValue(*p.Timezone)
	}
	if p.DiscordWebhookURL != nil {
		m.inputs[inputIndex(projSettingsDiscordURL)].SetValue(*p.DiscordWebhookURL)
	}
	if p.WeeklySummaryDay != nil {
		m.inputs[inputIndex(projSettingsWeeklySummaryDay)].SetValue(strconv.Itoa(*p.WeeklySummaryDay))
	} else {
		m.inputs[inputIndex(projSettingsWeeklySummaryDay)].SetValue("")
	}
	if p.WeeklySummaryTime != nil {
		m.inputs[inputIndex(projSettingsWeeklySummaryTime)].SetValue(*p.WeeklySummaryTime)
	} else {
		m.inputs[inputIndex(projSettingsWeeklySummaryTime)].SetValue("")
	}
	m.enabledToggle = p.SummaryEnabled
}

// renderInput renders a labelled input field with focus highlighting
func (m ProjectSettingsModel) renderInput(label string, focusConst int, b *strings.Builder) {
	b.WriteString(labelStyle.Render(label))
	b.WriteString("\n")
	idx := inputIndex(focusConst)
	if m.focusIndex == focusConst {
		b.WriteString(focusedInputStyle.Render(m.inputs[idx].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[idx].View()))
	}
	b.WriteString("\n\n")
}

// View renders the form
func (m ProjectSettingsModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Project Settings"))
	b.WriteString("\n\n")

	if m.loading && m.phase == "select" {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading projects..."))
		return baseStyle.Render(b.String())
	}

	if m.phase == "select" {
		b.WriteString(labelStyle.Render("Select a project to edit:"))
		b.WriteString("\n\n")

		for i, p := range m.projects {
			if i == m.selectedIdx {
				b.WriteString(fmt.Sprintf("  > %s\n", lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(p.Name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", p.Name))
			}
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Navigate  [Enter] Select  [Esc] Back"))
		return baseStyle.Render(b.String())
	}

	// Edit phase
	proj := m.projects[m.selectedIdx]
	b.WriteString(labelStyle.Render(fmt.Sprintf("Editing: %s", proj.Name)))
	b.WriteString("\n\n")

	m.renderInput("Project name:", projSettingsName, &b)

	// Summary enabled toggle
	b.WriteString(labelStyle.Render("Daily summary enabled:"))
	b.WriteString("\n")
	if m.focusIndex == projSettingsEnabled {
		if m.enabledToggle {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(successColor).Render("[x] Enabled"))
		} else {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("[ ] Disabled"))
		}
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  (space to toggle)"))
	} else {
		if m.enabledToggle {
			b.WriteString(lipgloss.NewStyle().Foreground(successColor).Render("[x] Enabled"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(errorColor).Render("[ ] Disabled"))
		}
	}
	b.WriteString("\n\n")

	m.renderInput("Daily summary time (HH:MM):", projSettingsSummaryTime, &b)
	m.renderInput("Timezone (IANA):", projSettingsTimezone, &b)
	m.renderInput("Discord webhook URL:", projSettingsDiscordURL, &b)

	// Weekly summary section
	b.WriteString(labelStyle.Render("Weekly attendance summary"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Sends a full-week attendance report to the same Discord channel."))
	b.WriteString("\n\n")

	m.renderInput("Day of week (1=Mon … 7=Sun, blank to disable):", projSettingsWeeklySummaryDay, &b)
	m.renderInput("Send time (HH:MM, blank to use daily summary time):", projSettingsWeeklySummaryTime, &b)

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
		b.WriteString(helpStyle.Render("[↑↓/Tab] Navigate  [Space] Toggle  [Enter/Ctrl+S] Save  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *ProjectSettingsModel) fetchProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectService.GetAll()
		if err != nil {
			return projectSettingsErrorMsg(err.Error())
		}
		return projectSettingsProjectsMsg{projects: projects}
	}
}

func (m *ProjectSettingsModel) handleSubmit() tea.Cmd {
	name := strings.TrimSpace(m.inputs[inputIndex(projSettingsName)].Value())
	summaryTime := strings.TrimSpace(m.inputs[inputIndex(projSettingsSummaryTime)].Value())
	timezone := strings.TrimSpace(m.inputs[inputIndex(projSettingsTimezone)].Value())
	discordURL := strings.TrimSpace(m.inputs[inputIndex(projSettingsDiscordURL)].Value())
	weeklySummaryDayStr := strings.TrimSpace(m.inputs[inputIndex(projSettingsWeeklySummaryDay)].Value())
	weeklySummaryTime := strings.TrimSpace(m.inputs[inputIndex(projSettingsWeeklySummaryTime)].Value())
	enabled := m.enabledToggle

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	projectID := m.projects[m.selectedIdx].ID

	return func() tea.Msg {
		req := &api.UpdateProjectRequest{
			SummaryEnabled: &enabled,
		}

		if name != "" {
			req.Name = &name
		}
		if summaryTime != "" {
			req.SummaryTime = &summaryTime
		}
		if timezone != "" {
			req.Timezone = &timezone
		}
		if discordURL != "" {
			req.DiscordWebhookURL = &discordURL
		}
		if weeklySummaryDayStr != "" {
			day, err := strconv.Atoi(weeklySummaryDayStr)
			if err != nil || day < 1 || day > 7 {
				return projectSettingsErrorMsg("Weekly summary day must be a number between 1 (Mon) and 7 (Sun)")
			}
			req.WeeklySummaryDay = &day
		}
		if weeklySummaryTime != "" {
			req.WeeklySummaryTime = &weeklySummaryTime
		}

		_, err := m.projectService.Update(projectID, req)
		if err != nil {
			return projectSettingsErrorMsg(err.Error())
		}
		return projectSettingsUpdatedMsg{}
	}
}

// Message types
type projectSettingsProjectsMsg struct {
	projects []*api.ProjectResponse
}
type projectSettingsUpdatedMsg struct{}
type projectSettingsErrorMsg string

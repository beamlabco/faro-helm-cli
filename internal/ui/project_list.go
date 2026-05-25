package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/project"
)

// ProjectListModel displays the user's projects
type ProjectListModel struct {
	projectService *project.Service
	projects       []*api.ProjectResponse
	loading        bool
	errorMsg       string
	shouldGoBack   bool
}

// NewProjectListModel creates a new project list model
func NewProjectListModel(projectService *project.Service) ProjectListModel {
	return ProjectListModel{
		projectService: projectService,
		loading:        true,
	}
}

// Init fetches projects
func (m ProjectListModel) Init() tea.Cmd {
	return m.fetchProjects()
}

// Update handles messages
func (m ProjectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			m.shouldGoBack = true
			return m, nil
		}

	case projectListLoadedMsg:
		m.loading = false
		m.projects = msg.projects
		return m, nil

	case projectListErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the project list
func (m ProjectListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Your Projects"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading projects..."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n\n")
	}

	if len(m.projects) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No projects found."))
		b.WriteString("\n")
	} else {
		for i, p := range m.projects {
			nameStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			dimStyle := lipgloss.NewStyle().Foreground(mutedColor)

			b.WriteString(fmt.Sprintf("  %d. %s", i+1, nameStyle.Render(p.Name)))

			var details []string
			if p.SummaryEnabled {
				details = append(details, "summary: on")
			}
			if p.SummaryTime != nil {
				details = append(details, fmt.Sprintf("time: %s", *p.SummaryTime))
			}
			if p.Timezone != nil {
				details = append(details, fmt.Sprintf("tz: %s", *p.Timezone))
			}
			if p.DiscordWebhookURL != nil {
				details = append(details, "discord: configured")
			}

			if len(details) > 0 {
				b.WriteString(dimStyle.Render(fmt.Sprintf(" (%s)", strings.Join(details, ", "))))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[esc/q] Back"))

	return baseStyle.Render(b.String())
}

func (m *ProjectListModel) fetchProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectService.GetMyProjects()
		if err != nil {
			return projectListErrorMsg(err.Error())
		}
		return projectListLoadedMsg{projects: projects}
	}
}

// Message types
type projectListLoadedMsg struct {
	projects []*api.ProjectResponse
}
type projectListErrorMsg string

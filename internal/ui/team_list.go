package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/team"
)

// TeamListModel displays the user's teams
type TeamListModel struct {
	teamService  *team.Service
	teams        []*api.TeamResponse
	loading      bool
	errorMsg     string
	shouldGoBack bool
}

// NewTeamListModel creates a new team list model
func NewTeamListModel(teamService *team.Service) TeamListModel {
	return TeamListModel{
		teamService: teamService,
		loading:     true,
	}
}

// Init fetches teams
func (m TeamListModel) Init() tea.Cmd {
	return m.fetchTeams()
}

// Update handles messages
func (m TeamListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			m.shouldGoBack = true
			return m, nil
		}

	case teamListLoadedMsg:
		m.loading = false
		m.teams = msg.teams
		return m, nil

	case teamListErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the team list
func (m TeamListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Your Teams"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading teams..."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n\n")
	}

	if len(m.teams) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No teams found."))
		b.WriteString("\n")
	} else {
		for i, t := range m.teams {
			nameStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			dimStyle := lipgloss.NewStyle().Foreground(mutedColor)

			b.WriteString(fmt.Sprintf("  %d. %s", i+1, nameStyle.Render(t.Name)))

			if t.Timezone != nil {
				b.WriteString(dimStyle.Render(fmt.Sprintf(" (tz: %s)", *t.Timezone)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("[esc/q] Back"))

	return baseStyle.Render(b.String())
}

func (m *TeamListModel) fetchTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := m.teamService.GetMyTeams()
		if err != nil {
			return teamListErrorMsg(err.Error())
		}
		return teamListLoadedMsg{teams: teams}
	}
}

// Message types
type teamListLoadedMsg struct {
	teams []*api.TeamResponse
}
type teamListErrorMsg string

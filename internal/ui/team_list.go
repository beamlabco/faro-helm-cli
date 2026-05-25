package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/user"
)

// TeamListModel represents the team members list view
type TeamListModel struct {
	userService  *user.Service
	members      []*api.MemberResponse
	errorMsg     string
	loading      bool
	shouldGoBack bool
}

// NewTeamListModel creates a new team list model
func NewTeamListModel(userService *user.Service) TeamListModel {
	return TeamListModel{
		userService: userService,
		loading:     true,
	}
}

// Init initializes the model
func (m TeamListModel) Init() tea.Cmd {
	return m.fetchMembers()
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

		case "r":
			m.loading = true
			m.errorMsg = ""
			return m, m.fetchMembers()
		}

	case teamListSuccessMsg:
		m.loading = false
		m.members = msg.members
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

	b.WriteString(titleStyle.Render("Team Members"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading team members..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.members) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No team members found"))
		b.WriteString("\n")
	} else {
		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
		emailStyle := lipgloss.NewStyle().Foreground(mutedColor)

		for i, member := range m.members {
			if i > 0 {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", 50)))
				b.WriteString("\n")
			}

			b.WriteString(nameStyle.Render(member.Name))
			b.WriteString("  ")
			b.WriteString(getRoleBadgeStyle(member.Role).Render(member.Role))
			b.WriteString("\n")
			b.WriteString(emailStyle.Render("  "+member.Email))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("%d members", len(m.members))))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("[r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *TeamListModel) fetchMembers() tea.Cmd {
	return func() tea.Msg {
		members, err := m.userService.GetMembers()
		if err != nil {
			return teamListErrorMsg(err.Error())
		}
		return teamListSuccessMsg{members: members}
	}
}

// Message types
type teamListSuccessMsg struct {
	members []*api.MemberResponse
}
type teamListErrorMsg string

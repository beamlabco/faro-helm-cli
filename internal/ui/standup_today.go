package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/standup"
)

// StandupTodayModel represents the current user's today's standups view
type StandupTodayModel struct {
	standupService *standup.Service
	standups       []*api.StandupResponse
	errorMsg       string
	loading        bool
	shouldGoBack   bool
	onBack         func()
}

// NewStandupTodayModel creates a new standup today model
func NewStandupTodayModel(standupService *standup.Service, onBack func()) StandupTodayModel {
	return StandupTodayModel{
		standupService: standupService,
		loading:        true,
		onBack:         onBack,
	}
}

// Init initializes the standup today model
func (m StandupTodayModel) Init() tea.Cmd {
	return m.fetchStandups()
}

// Update handles messages
func (m StandupTodayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, m.fetchStandups()
		}

	case standupTodaySuccessMsg:
		m.loading = false
		m.standups = msg.standups
		return m, nil

	case standupTodayErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the today's standups view
func (m StandupTodayModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Your Standup Today"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading standup..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.standups) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("You haven't submitted a standup today."))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Use /standup to submit it."))
		b.WriteString("\n")
	} else {
		for i, s := range m.standups {
			if i > 0 {
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", 60)))
				b.WriteString("\n\n")
			}

			if s.Team != nil {
				teamStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
				b.WriteString(teamStyle.Render(s.Team.Name))
				b.WriteString("\n\n")
			}

			if s.Yesterday != nil && *s.Yesterday != "" {
				b.WriteString(labelStyle.Render("Yesterday:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(*s.Yesterday))
				b.WriteString("\n\n")
			}

			if s.Today != nil && *s.Today != "" {
				b.WriteString(labelStyle.Render("Today:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(*s.Today))
				b.WriteString("\n\n")
			}

			if s.Blockers != nil && *s.Blockers != "" {
				b.WriteString(labelStyle.Render("Blockers:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(errorColor).Render(*s.Blockers))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("r: refresh • esc: back"))
	}

	return baseStyle.Render(b.String())
}

// fetchStandups fetches the current user's standups for today
func (m *StandupTodayModel) fetchStandups() tea.Cmd {
	return func() tea.Msg {
		standups, err := m.standupService.GetToday()
		if err != nil {
			return standupTodayErrorMsg(err.Error())
		}
		return standupTodaySuccessMsg{standups: standups}
	}
}

// Message types
type standupTodaySuccessMsg struct {
	standups []*api.StandupResponse
}
type standupTodayErrorMsg string

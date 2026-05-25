package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/standup"
)

// StandupTodayModel represents the today's standups view
type StandupTodayModel struct {
	standupService *standup.Service
	standups       []*api.StandupResponse
	date           string
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
		m.date = msg.date
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

	// Title
	b.WriteString(titleStyle.Render(fmt.Sprintf("👥 Team Standups - %s", m.date)))
	b.WriteString("\n\n")

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("⏳ Loading standups..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		// Error message
		b.WriteString(errorStyle.Render("❌ " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.standups) == 0 {
		// No standups
		b.WriteString(lipgloss.NewStyle().
			Foreground(mutedColor).
			Render("No standups submitted yet today"))
		b.WriteString("\n")
	} else {
		// Display standups
		for i, standup := range m.standups {
			if i > 0 {
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().
					Foreground(mutedColor).
					Render(strings.Repeat("─", 60)))
				b.WriteString("\n\n")
			}

			// User name
			userName := "Unknown"
			userRole := ""
			if standup.User != nil {
				userName = standup.User.Name
				userRole = standup.User.Role
			}

			nameStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

			roleStyle := lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

			b.WriteString(nameStyle.Render(userName))
			b.WriteString(" ")
			b.WriteString(roleStyle.Render(fmt.Sprintf("(%s)", userRole)))
			b.WriteString("\n\n")

			// Yesterday
			if standup.Yesterday != nil && *standup.Yesterday != "" {
				b.WriteString(labelStyle.Render("Yesterday:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Render(*standup.Yesterday))
				b.WriteString("\n\n")
			}

			// Today
			if standup.Today != nil && *standup.Today != "" {
				b.WriteString(labelStyle.Render("Today:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFFFF")).
					Render(*standup.Today))
				b.WriteString("\n\n")
			}

			// Blockers
			if standup.Blockers != nil && *standup.Blockers != "" {
				b.WriteString(labelStyle.Render("Blockers:"))
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().
					Foreground(errorColor).
					Render(*standup.Blockers))
				b.WriteString("\n")
			}
		}
	}

	// Help text
	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("r: refresh • esc: back"))
	}

	return baseStyle.Render(b.String())
}

// fetchStandups fetches today's standups
func (m *StandupTodayModel) fetchStandups() tea.Cmd {
	return func() tea.Msg {
		standups, date, err := m.standupService.GetToday()
		if err != nil {
			return standupTodayErrorMsg(err.Error())
		}
		return standupTodaySuccessMsg{
			standups: standups,
			date:     date,
		}
	}
}

// Message types
type standupTodaySuccessMsg struct {
	standups []*api.StandupResponse
	date     string
}
type standupTodayErrorMsg string

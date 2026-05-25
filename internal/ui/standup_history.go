package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/standup"
)

// StandupHistoryModel represents the my standups history view
type StandupHistoryModel struct {
	standupService *standup.Service
	standups       []*api.StandupResponse
	pagination     *api.Pagination
	errorMsg       string
	loading        bool
	shouldGoBack   bool
	onBack         func()
}

// NewStandupHistoryModel creates a new standup history model
func NewStandupHistoryModel(standupService *standup.Service, onBack func()) StandupHistoryModel {
	return StandupHistoryModel{
		standupService: standupService,
		loading:        true,
		onBack:         onBack,
	}
}

// Init initializes the standup history model
func (m StandupHistoryModel) Init() tea.Cmd {
	return m.fetchStandups()
}

// Update handles messages
func (m StandupHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case standupHistorySuccessMsg:
		m.loading = false
		m.standups = msg.standups
		m.pagination = msg.pagination
		return m, nil

	case standupHistoryErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the my standups history view
func (m StandupHistoryModel) View() string {

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📋 My Standup History"))
	b.WriteString("\n\n")

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("⏳ Loading your standups..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		// Error message
		b.WriteString(errorStyle.Render("❌ " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.standups) == 0 {
		// No standups
		b.WriteString(lipgloss.NewStyle().
			Foreground(mutedColor).
			Render("You haven't submitted any standups yet"))
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

			// Date
			dateStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

			b.WriteString(dateStyle.Render(standup.Date))
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

		// Pagination info
		if m.pagination != nil {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(mutedColor).
				Render(fmt.Sprintf("Showing %d of %d standups", len(m.standups), m.pagination.Total)))
			b.WriteString("\n")
		}
	}

	// Help text
	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("r: refresh • esc: back"))
	}

	return baseStyle.Render(b.String())
}

// fetchStandups fetches the user's standup history
func (m *StandupHistoryModel) fetchStandups() tea.Cmd {
	return func() tea.Msg {
		standups, pagination, err := m.standupService.GetMy(10, 0)
		if err != nil {
			return standupHistoryErrorMsg(err.Error())
		}
		return standupHistorySuccessMsg{
			standups:   standups,
			pagination: pagination,
		}
	}
}

// Message types
type standupHistorySuccessMsg struct {
	standups   []*api.StandupResponse
	pagination *api.Pagination
}
type standupHistoryErrorMsg string

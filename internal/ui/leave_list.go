package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/leave"
)

// LeaveListModel represents the leave list view
type LeaveListModel struct {
	leaveService *leave.Service
	leaves       []*api.LeaveResponse
	count        int
	errorMsg     string
	loading      bool
	shouldGoBack bool
}

// NewLeaveListModel creates a new leave list model
func NewLeaveListModel(leaveService *leave.Service) LeaveListModel {
	return LeaveListModel{
		leaveService: leaveService,
		loading:      true,
	}
}

// Init initializes the model
func (m LeaveListModel) Init() tea.Cmd {
	return m.fetchLeaves()
}

// Update handles messages
func (m LeaveListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, m.fetchLeaves()
		}

	case leaveListSuccessMsg:
		m.loading = false
		m.leaves = msg.leaves
		m.count = msg.count
		return m, nil

	case leaveListErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the leave list
func (m LeaveListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Team Leaves"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading leaves..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.leaves) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No leave requests found"))
		b.WriteString("\n")
	} else {
		for i, l := range m.leaves {
			if i > 0 {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", 55)))
				b.WriteString("\n")
			}

			// User name + leave type
			userName := "Unknown"
			if l.User != nil {
				userName = l.User.Name
			}
			nameStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			b.WriteString(nameStyle.Render(userName))
			b.WriteString("  ")

			typeName := "Unknown"
			if l.Type != nil {
				typeName = l.Type.Name
			}
			typeStyle := lipgloss.NewStyle().Foreground(secondaryColor)
			b.WriteString(typeStyle.Render(fmt.Sprintf("%s %s", getLeaveTypeIcon(typeName), typeName)))
			b.WriteString("\n")

			// Dates
			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			b.WriteString(dateStyle.Render(fmt.Sprintf("  %s → %s", l.StartDate, l.EndDate)))
			b.WriteString("  ")

			// Status
			statusStyle := getLeaveStatusStyle(l.Status)
			b.WriteString(statusStyle.Render(fmt.Sprintf("[%s]", l.Status)))
			b.WriteString("\n")

			// Reason
			if l.Reason != nil && *l.Reason != "" {
				reasonStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
				b.WriteString(reasonStyle.Render("  " + *l.Reason))
				b.WriteString("\n")
			}

			// Reviewer
			if l.ReviewedBy != nil {
				reviewerStyle := lipgloss.NewStyle().Foreground(mutedColor)
				b.WriteString(reviewerStyle.Render(fmt.Sprintf("  Reviewed by: %s", *l.ReviewedBy)))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("Showing %d of %d leaves", len(m.leaves), m.count)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("[r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *LeaveListModel) fetchLeaves() tea.Cmd {
	return func() tea.Msg {
		leaves, count, err := m.leaveService.GetAll("", "", 20, 0)
		if err != nil {
			return leaveListErrorMsg(err.Error())
		}
		return leaveListSuccessMsg{leaves: leaves, count: count}
	}
}

func getLeaveStatusStyle(status string) lipgloss.Style {
	switch status {
	case "approved":
		return lipgloss.NewStyle().Foreground(successColor)
	case "rejected":
		return lipgloss.NewStyle().Foreground(errorColor)
	case "pending":
		return lipgloss.NewStyle().Foreground(secondaryColor)
	default:
		return lipgloss.NewStyle().Foreground(mutedColor)
	}
}

// Message types
type leaveListSuccessMsg struct {
	leaves []*api.LeaveResponse
	count  int
}
type leaveListErrorMsg string

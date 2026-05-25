package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/leave"
)

// LeaveReviewModel represents the leave review view (pending leaves for managers/primary)
type LeaveReviewModel struct {
	leaveService  *leave.Service
	leaves        []*api.LeaveResponse
	count         int
	selectedIndex int
	errorMsg      string
	successMsg    string
	loading       bool
	shouldGoBack  bool
}

// NewLeaveReviewModel creates a new leave review model
func NewLeaveReviewModel(leaveService *leave.Service) LeaveReviewModel {
	return LeaveReviewModel{
		leaveService: leaveService,
		loading:      true,
	}
}

// Init initializes the model
func (m LeaveReviewModel) Init() tea.Cmd {
	return m.fetchPendingLeaves()
}

// Update handles messages
func (m LeaveReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc", "q":
			m.shouldGoBack = true
			return m, nil

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.leaves)-1 {
				m.selectedIndex++
			}
			return m, nil

		case "a":
			if !m.loading && len(m.leaves) > 0 {
				return m, m.handleAction("approved")
			}

		case "x":
			if !m.loading && len(m.leaves) > 0 {
				return m, m.handleAction("rejected")
			}

		case "r":
			m.loading = true
			m.errorMsg = ""
			m.successMsg = ""
			return m, m.fetchPendingLeaves()
		}

	case leaveReviewFetchSuccessMsg:
		m.loading = false
		m.leaves = msg.leaves
		m.count = msg.count
		if m.selectedIndex >= len(m.leaves) {
			m.selectedIndex = max(0, len(m.leaves)-1)
		}
		return m, nil

	case leaveReviewActionSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Leave %s!", msg.status)
		// Refresh the list
		return m, m.fetchPendingLeaves()

	case leaveReviewErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the leave review view
func (m LeaveReviewModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Review Pending Leaves"))
	b.WriteString("\n\n")

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading pending leaves..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.leaves) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No pending leave requests"))
		b.WriteString("\n")
	} else {
		for i, l := range m.leaves {
			// Selection indicator
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "▸ "
			}

			userName := "Unknown"
			if l.User != nil {
				userName = l.User.Name
			}

			var nameStyle lipgloss.Style
			if i == m.selectedIndex {
				nameStyle = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			} else {
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			}

			b.WriteString(prefix)
			b.WriteString(nameStyle.Render(userName))
			b.WriteString("  ")

			typeStyle := lipgloss.NewStyle().Foreground(secondaryColor)
			b.WriteString(typeStyle.Render(fmt.Sprintf("%s %s", getLeaveTypeIcon(l.Type), l.Type)))
			b.WriteString("\n")

			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			b.WriteString("    ")
			b.WriteString(dateStyle.Render(fmt.Sprintf("%s → %s", l.StartDate, l.EndDate)))
			b.WriteString("\n")

			if l.Reason != nil && *l.Reason != "" {
				reasonStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
				b.WriteString("    ")
				b.WriteString(reasonStyle.Render(*l.Reason))
				b.WriteString("\n")
			}

			if i < len(m.leaves)-1 {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  " + strings.Repeat("─", 50)))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("%d pending", m.count)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("[↑↓] Select  [a] Approve  [x] Reject  [r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *LeaveReviewModel) fetchPendingLeaves() tea.Cmd {
	return func() tea.Msg {
		leaves, count, err := m.leaveService.GetAll("pending", "", 20, 0)
		if err != nil {
			return leaveReviewErrorMsg(err.Error())
		}
		return leaveReviewFetchSuccessMsg{leaves: leaves, count: count}
	}
}

func (m *LeaveReviewModel) handleAction(status string) tea.Cmd {
	selected := m.leaves[m.selectedIndex]
	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		_, err := m.leaveService.UpdateStatus(selected.ID, status)
		if err != nil {
			return leaveReviewErrorMsg(err.Error())
		}
		return leaveReviewActionSuccessMsg{status: status}
	}
}

// Message types
type leaveReviewFetchSuccessMsg struct {
	leaves []*api.LeaveResponse
	count  int
}
type leaveReviewActionSuccessMsg struct {
	status string
}
type leaveReviewErrorMsg string

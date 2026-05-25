package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/leave"
)

// LeaveCancelModel represents the leave cancellation view (user's own pending leaves)
type LeaveCancelModel struct {
	leaveService  *leave.Service
	userID        string
	leaves        []*api.LeaveResponse
	count         int
	selectedIndex int
	errorMsg      string
	successMsg    string
	loading       bool
	shouldGoBack  bool
}

// NewLeaveCancelModel creates a new leave cancel model
func NewLeaveCancelModel(leaveService *leave.Service, userID string) LeaveCancelModel {
	return LeaveCancelModel{
		leaveService: leaveService,
		userID:       userID,
		loading:      true,
	}
}

// Init initializes the model
func (m LeaveCancelModel) Init() tea.Cmd {
	return m.fetchMyPendingLeaves()
}

// Update handles messages
func (m LeaveCancelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case "d":
			if !m.loading && len(m.leaves) > 0 {
				return m, m.handleCancel()
			}

		case "r":
			m.loading = true
			m.errorMsg = ""
			m.successMsg = ""
			return m, m.fetchMyPendingLeaves()
		}

	case leaveCancelFetchSuccessMsg:
		m.loading = false
		m.leaves = msg.leaves
		m.count = msg.count
		if m.selectedIndex >= len(m.leaves) {
			m.selectedIndex = max(0, len(m.leaves)-1)
		}
		return m, nil

	case leaveCancelActionSuccessMsg:
		m.loading = false
		m.successMsg = "Leave request cancelled!"
		return m, m.fetchMyPendingLeaves()

	case leaveCancelErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the cancel leave view
func (m LeaveCancelModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Cancel Leave Request"))
	b.WriteString("\n\n")

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading your pending leaves..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.leaves) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("You have no pending leave requests"))
		b.WriteString("\n")
	} else {
		for i, l := range m.leaves {
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "▸ "
			}

			var typeStyle lipgloss.Style
			if i == m.selectedIndex {
				typeStyle = lipgloss.NewStyle().Bold(true).Foreground(secondaryColor)
			} else {
				typeStyle = lipgloss.NewStyle().Foreground(secondaryColor)
			}

			b.WriteString(prefix)
			b.WriteString(typeStyle.Render(fmt.Sprintf("%s %s", getLeaveTypeIcon(l.Type), l.Type)))
			b.WriteString("  ")

			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
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
		b.WriteString(helpStyle.Render("[↑↓] Select  [d] Cancel leave  [r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *LeaveCancelModel) fetchMyPendingLeaves() tea.Cmd {
	return func() tea.Msg {
		leaves, count, err := m.leaveService.GetAll("pending", m.userID, 20, 0)
		if err != nil {
			return leaveCancelErrorMsg(err.Error())
		}
		return leaveCancelFetchSuccessMsg{leaves: leaves, count: count}
	}
}

func (m *LeaveCancelModel) handleCancel() tea.Cmd {
	selected := m.leaves[m.selectedIndex]
	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		err := m.leaveService.Cancel(selected.ID)
		if err != nil {
			return leaveCancelErrorMsg(err.Error())
		}
		return leaveCancelActionSuccessMsg{}
	}
}

// Message types
type leaveCancelFetchSuccessMsg struct {
	leaves []*api.LeaveResponse
	count  int
}
type leaveCancelActionSuccessMsg struct{}
type leaveCancelErrorMsg string

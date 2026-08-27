package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/leave"
)

// LeaveRequestModel represents the leave request form
type LeaveRequestModel struct {
	leaveService   *leave.Service
	balances       []*api.LeaveTypeBalance
	selectedType   int
	startDateInput textinput.Model
	endDateInput   textinput.Model
	reasonInput    textinput.Model
	focusIndex     int    // 1=startDate, 2=endDate, 3=reason
	phase          string // "loading", "type_select", "form"
	errorMsg       string
	successMsg     string
	loading        bool
	shouldGoBack   bool
}

// NewLeaveRequestModel creates a new leave request model
func NewLeaveRequestModel(leaveService *leave.Service) LeaveRequestModel {
	startDate := textinput.New()
	startDate.Placeholder = "YYYY-MM-DD"
	startDate.CharLimit = 10
	startDate.Width = 20
	startDate.SetValue(time.Now().Format("2006-01-02"))

	endDate := textinput.New()
	endDate.Placeholder = "YYYY-MM-DD"
	endDate.CharLimit = 10
	endDate.Width = 20
	endDate.SetValue(time.Now().Format("2006-01-02"))

	reason := textinput.New()
	reason.Placeholder = "Optional reason..."
	reason.CharLimit = 500
	reason.Width = 50

	return LeaveRequestModel{
		leaveService:   leaveService,
		startDateInput: startDate,
		endDateInput:   endDate,
		reasonInput:    reason,
		focusIndex:     1,
		phase:          "loading",
		loading:        true,
	}
}

// Init fetches the workspace's leave types and the user's balance for each
func (m LeaveRequestModel) Init() tea.Cmd {
	return m.fetchBalances()
}

func (m *LeaveRequestModel) fetchBalances() tea.Cmd {
	return func() tea.Msg {
		balances, _, err := m.leaveService.GetBalance(0)
		if err != nil {
			return leaveRequestErrorMsg(err.Error())
		}
		return leaveBalancesLoadedMsg{balances: balances}
	}
}

// Update handles messages
func (m LeaveRequestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.phase == "form" && len(m.balances) > 1 {
				m.phase = "type_select"
				return m, nil
			}
			m.shouldGoBack = true
			return m, nil
		}

		if m.phase == "type_select" {
			switch msg.String() {
			case "left", "h", "up", "k":
				m.selectedType--
				if m.selectedType < 0 {
					m.selectedType = len(m.balances) - 1
				}
				return m, nil

			case "right", "l", "down", "j":
				m.selectedType++
				if m.selectedType >= len(m.balances) {
					m.selectedType = 0
				}
				return m, nil

			case "enter":
				if len(m.balances) == 0 {
					return m, nil
				}
				m.phase = "form"
				m.focusIndex = 1
				m.focusCurrent()
				return m, nil
			}
			return m, nil
		}

		if m.phase == "form" {
			switch msg.String() {
			case "tab", "down", "j":
				m.blurAll()
				m.focusIndex++
				if m.focusIndex > 3 {
					m.focusIndex = 1
				}
				m.focusCurrent()
				return m, nil

			case "shift+tab", "up", "k":
				m.blurAll()
				m.focusIndex--
				if m.focusIndex < 1 {
					m.focusIndex = 3
				}
				m.focusCurrent()
				return m, nil

			case "enter":
				if m.loading {
					return m, nil
				}
				return m, m.handleSubmit()
			}
		}

	case leaveBalancesLoadedMsg:
		m.loading = false
		m.balances = msg.balances
		if len(m.balances) == 0 {
			m.errorMsg = "No leave types configured for your workspace. Contact your admin."
			return m, nil
		}
		if len(m.balances) == 1 {
			m.phase = "form"
			m.focusIndex = 1
			m.focusCurrent()
		} else {
			m.phase = "type_select"
		}
		return m, nil

	case leaveRequestSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Leave request created! (%s: %s to %s)", msg.leaveType, msg.startDate, msg.endDate)
		return m, nil

	case leaveRequestErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	if m.phase != "form" {
		return m, nil
	}

	// Update focused input
	var cmd tea.Cmd
	switch m.focusIndex {
	case 1:
		m.startDateInput, cmd = m.startDateInput.Update(msg)
	case 2:
		m.endDateInput, cmd = m.endDateInput.Update(msg)
	case 3:
		m.reasonInput, cmd = m.reasonInput.Update(msg)
	}
	return m, cmd
}

func (m *LeaveRequestModel) blurAll() {
	m.startDateInput.Blur()
	m.endDateInput.Blur()
	m.reasonInput.Blur()
}

func (m *LeaveRequestModel) focusCurrent() {
	m.blurAll()
	switch m.focusIndex {
	case 1:
		m.startDateInput.Focus()
	case 2:
		m.endDateInput.Focus()
	case 3:
		m.reasonInput.Focus()
	}
}

// View renders the leave request form
func (m LeaveRequestModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Request Leave"))
	b.WriteString("\n\n")

	if m.phase == "loading" {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading leave types..."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	if len(m.balances) == 0 {
		if m.errorMsg != "" {
			b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
			b.WriteString("\n\n")
		}
		b.WriteString(helpStyle.Render("[Esc] Back"))
		return baseStyle.Render(b.String())
	}

	if m.phase == "type_select" {
		b.WriteString(labelStyle.Render("Leave type:"))
		b.WriteString("\n\n")

		for i, bal := range m.balances {
			prefix := "  "
			var style lipgloss.Style
			if i == m.selectedType {
				prefix = "▸ "
				style = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			} else {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			}

			b.WriteString(prefix)
			b.WriteString(style.Render(fmt.Sprintf("%s %s", getLeaveTypeIcon(bal.Type), bal.Type)))
			b.WriteString("  ")
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(formatLeaveBalance(bal)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Select  [Enter] Continue  [Esc] Back"))
		return baseStyle.Render(b.String())
	}

	// Form phase
	selected := m.balances[m.selectedType]
	b.WriteString(labelStyle.Render("Leave type:"))
	b.WriteString(" ")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(fmt.Sprintf("%s %s", getLeaveTypeIcon(selected.Type), selected.Type)))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(formatLeaveBalance(selected)))
	b.WriteString("\n\n")

	// Start date
	b.WriteString(labelStyle.Render("Start date:"))
	b.WriteString("\n")
	if m.focusIndex == 1 {
		b.WriteString(focusedInputStyle.Render(m.startDateInput.View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.startDateInput.View()))
	}
	b.WriteString("\n\n")

	// End date
	b.WriteString(labelStyle.Render("End date:"))
	b.WriteString("\n")
	if m.focusIndex == 2 {
		b.WriteString(focusedInputStyle.Render(m.endDateInput.View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.endDateInput.View()))
	}
	b.WriteString("\n\n")

	// Reason
	b.WriteString(labelStyle.Render("Reason (optional):"))
	b.WriteString("\n")
	if m.focusIndex == 3 {
		b.WriteString(focusedInputStyle.Render(m.reasonInput.View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.reasonInput.View()))
	}
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Submitting leave request..."))
		b.WriteString("\n")
	}

	if !m.loading {
		help := "[↑↓/Tab] Navigate  [Enter] Submit  [Esc] Back"
		if len(m.balances) > 1 {
			help = "[↑↓/Tab] Navigate  [Enter] Submit  [Esc] Change type"
		}
		b.WriteString(helpStyle.Render(help))
	}

	return baseStyle.Render(b.String())
}

func (m *LeaveRequestModel) handleSubmit() tea.Cmd {
	selected := m.balances[m.selectedType]
	startDate := strings.TrimSpace(m.startDateInput.Value())
	endDate := strings.TrimSpace(m.endDateInput.Value())
	reason := strings.TrimSpace(m.reasonInput.Value())
	leaveType := selected.Type

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		_, err := m.leaveService.Create(selected.TypeID, startDate, endDate, reason)
		if err != nil {
			return leaveRequestErrorMsg(err.Error())
		}
		return leaveRequestSuccessMsg{leaveType: leaveType, startDate: startDate, endDate: endDate}
	}
}

// formatLeaveBalance renders a leave type's quota/used/remaining for display
func formatLeaveBalance(b *api.LeaveTypeBalance) string {
	if b.Quota == nil {
		return fmt.Sprintf("%d used, unlimited", b.Used)
	}
	remaining := *b.Quota - b.Used
	if b.Remaining != nil {
		remaining = *b.Remaining
	}
	return fmt.Sprintf("%d/%d remaining", remaining, *b.Quota)
}

func getLeaveTypeIcon(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "sick":
		return "🤒"
	case "casual":
		return "🏖"
	case "paid":
		return "💰"
	case "unpaid":
		return "📋"
	case "wfh", "work from home", "remote":
		return "🏠"
	default:
		return "•"
	}
}

// Message types
type leaveBalancesLoadedMsg struct {
	balances []*api.LeaveTypeBalance
}
type leaveRequestSuccessMsg struct {
	leaveType string
	startDate string
	endDate   string
}
type leaveRequestErrorMsg string

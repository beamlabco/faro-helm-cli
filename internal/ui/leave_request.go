package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/leave"
)

// LeaveRequestModel represents the leave request form
type LeaveRequestModel struct {
	leaveService   *leave.Service
	typeOptions    []string
	selectedType   int
	startDateInput textinput.Model
	endDateInput   textinput.Model
	reasonInput    textinput.Model
	focusIndex     int // 0=type, 1=startDate, 2=endDate, 3=reason
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
		typeOptions:    []string{"sick", "casual", "paid", "unpaid", "wfh"},
		selectedType:   0,
		startDateInput: startDate,
		endDateInput:   endDate,
		reasonInput:    reason,
		focusIndex:     0,
	}
}

// Init initializes the model
func (m LeaveRequestModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m LeaveRequestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "left", "h":
			if m.focusIndex == 0 {
				m.selectedType--
				if m.selectedType < 0 {
					m.selectedType = len(m.typeOptions) - 1
				}
				return m, nil
			}

		case "right", "l":
			if m.focusIndex == 0 {
				m.selectedType++
				if m.selectedType >= len(m.typeOptions) {
					m.selectedType = 0
				}
				return m, nil
			}

		case "tab", "down", "j":
			m.blurAll()
			m.focusIndex++
			if m.focusIndex > 3 {
				m.focusIndex = 0
			}
			m.focusCurrent()
			return m, nil

		case "shift+tab", "up", "k":
			m.blurAll()
			m.focusIndex--
			if m.focusIndex < 0 {
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

	case leaveRequestSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Leave request created! (%s: %s to %s)", msg.leaveType, msg.startDate, msg.endDate)
		return m, nil

	case leaveRequestErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
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

	// Type selection
	b.WriteString(labelStyle.Render("Leave type:"))
	b.WriteString("\n\n")

	for i, t := range m.typeOptions {
		var style lipgloss.Style
		if i == m.selectedType {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Padding(0, 2).
				MarginRight(1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(mutedColor).
				Padding(0, 2).
				MarginRight(1)
		}
		icon := getLeaveTypeIcon(t)
		b.WriteString(style.Render(fmt.Sprintf("%s %s", icon, t)))
	}
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
		b.WriteString(helpStyle.Render("[←→] Type  [↑↓/Tab] Navigate  [Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *LeaveRequestModel) handleSubmit() tea.Cmd {
	leaveType := m.typeOptions[m.selectedType]
	startDate := strings.TrimSpace(m.startDateInput.Value())
	endDate := strings.TrimSpace(m.endDateInput.Value())
	reason := strings.TrimSpace(m.reasonInput.Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		_, err := m.leaveService.Create(leaveType, startDate, endDate, reason)
		if err != nil {
			return leaveRequestErrorMsg(err.Error())
		}
		return leaveRequestSuccessMsg{leaveType: leaveType, startDate: startDate, endDate: endDate}
	}
}

func getLeaveTypeIcon(t string) string {
	switch t {
	case "sick":
		return "🤒"
	case "casual":
		return "🏖"
	case "paid":
		return "💰"
	case "unpaid":
		return "📋"
	case "wfh":
		return "🏠"
	default:
		return "•"
	}
}

// Message types
type leaveRequestSuccessMsg struct {
	leaveType string
	startDate string
	endDate   string
}
type leaveRequestErrorMsg string

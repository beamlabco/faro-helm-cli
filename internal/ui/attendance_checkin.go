package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/attendance"
)

// AttendanceCheckInModel represents the check-in form
type AttendanceCheckInModel struct {
	attendanceService *attendance.Service
	notesInput        textinput.Model
	statusOptions     []string
	selectedStatus    int
	errorMsg          string
	successMsg        string
	loading           bool
	shouldGoBack      bool
}

// NewAttendanceCheckInModel creates a new check-in model
func NewAttendanceCheckInModel(attendanceService *attendance.Service) AttendanceCheckInModel {
	notes := textinput.New()
	notes.Placeholder = "Optional notes..."
	notes.CharLimit = 500
	notes.Width = 50

	notes.Blur()

	return AttendanceCheckInModel{
		attendanceService: attendanceService,
		notesInput:        notes,
		statusOptions:     []string{"present", "remote"},
		selectedStatus:    0,
	}
}

// Init initializes the check-in model
func (m AttendanceCheckInModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m AttendanceCheckInModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkInSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Checked in as %s at %s", msg.status, msg.time)
		return m, nil

	case checkInErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "tab":
			if m.notesInput.Focused() {
				m.notesInput.Blur()
			} else {
				m.notesInput.Focus()
			}
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			return m, m.handleSubmit()
		}

		// When notes input is focused, pass keys to it
		if m.notesInput.Focused() {
			var cmd tea.Cmd
			m.notesInput, cmd = m.notesInput.Update(msg)
			return m, cmd
		}

		// When notes input is NOT focused, handle status navigation
		switch msg.String() {
		case "left", "h":
			m.selectedStatus--
			if m.selectedStatus < 0 {
				m.selectedStatus = len(m.statusOptions) - 1
			}
			return m, nil

		case "right", "l":
			m.selectedStatus++
			if m.selectedStatus >= len(m.statusOptions) {
				m.selectedStatus = 0
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the check-in form
func (m AttendanceCheckInModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Check In"))
	b.WriteString("\n\n")

	// Status selection
	b.WriteString(labelStyle.Render("How are you working today?"))
	b.WriteString("\n\n")

	for i, status := range m.statusOptions {
		var style lipgloss.Style
		if i == m.selectedStatus {
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

		icon := getStatusIcon(status)
		b.WriteString(style.Render(fmt.Sprintf("%s %s", icon, status)))
	}
	b.WriteString("\n\n")

	// Notes input
	b.WriteString(labelStyle.Render("Notes (optional):"))
	b.WriteString("\n")
	if m.notesInput.Focused() {
		b.WriteString(focusedInputStyle.Render(m.notesInput.View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.notesInput.View()))
	}
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	// Success message
	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Checking in..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		b.WriteString(helpStyle.Render("[←→] Select status  [Tab] Notes  [Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

// handleSubmit handles the check-in form submission
func (m *AttendanceCheckInModel) handleSubmit() tea.Cmd {
	status := m.statusOptions[m.selectedStatus]
	notes := strings.TrimSpace(m.notesInput.Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		resp, err := m.attendanceService.CheckIn(status, notes)
		if err != nil {
			return checkInErrorMsg(err.Error())
		}
		t := ""
		if resp.CheckinAt != nil {
			t = utcTimeToLocal(*resp.CheckinAt)
		}
		return checkInSuccessMsg{status: status, time: t}
	}
}

// Message types
type checkInSuccessMsg struct {
	status string
	time   string
}
type checkInErrorMsg string

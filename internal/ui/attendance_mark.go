package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/attendance"
)

// AttendanceMarkModel represents the attendance marking form
type AttendanceMarkModel struct {
	attendanceService *attendance.Service
	notesInput        textinput.Model
	statusOptions     []string
	selectedStatus    int
	date              string
	errorMsg          string
	successMsg        string
	loading           bool
	shouldGoBack      bool
}

// NewAttendanceMarkModel creates a new attendance mark model
func NewAttendanceMarkModel(attendanceService *attendance.Service) AttendanceMarkModel {
	notes := textinput.New()
	notes.Placeholder = "Optional notes..."
	notes.CharLimit = 500
	notes.Width = 50

	return AttendanceMarkModel{
		attendanceService: attendanceService,
		notesInput:        notes,
		statusOptions:     []string{"present", "remote", "half-day", "absent"},
		selectedStatus:    0,
		date:              time.Now().Format("2006-01-02"),
	}
}

// Init initializes the attendance mark model
func (m AttendanceMarkModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m AttendanceMarkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "left", "h":
			if !m.notesInput.Focused() {
				m.selectedStatus--
				if m.selectedStatus < 0 {
					m.selectedStatus = len(m.statusOptions) - 1
				}
			}
			return m, nil

		case "right", "l":
			if !m.notesInput.Focused() {
				m.selectedStatus++
				if m.selectedStatus >= len(m.statusOptions) {
					m.selectedStatus = 0
				}
			}
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

	case attendanceMarkSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Marked as %s!", msg.status)
		return m, nil

	case attendanceMarkErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update notes input if focused
	if m.notesInput.Focused() {
		var cmd tea.Cmd
		m.notesInput, cmd = m.notesInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the attendance marking form
func (m AttendanceMarkModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(fmt.Sprintf("Mark Attendance - %s", m.date)))
	b.WriteString("\n\n")

	// Status selection
	b.WriteString(labelStyle.Render("Select your status:"))
	b.WriteString("\n\n")

	// Render status options as buttons
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
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Marking attendance..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		b.WriteString(helpStyle.Render("[←→] Select status  [Tab] Notes  [Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

// handleSubmit handles the attendance form submission
func (m *AttendanceMarkModel) handleSubmit() tea.Cmd {
	status := m.statusOptions[m.selectedStatus]
	notes := strings.TrimSpace(m.notesInput.Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		_, err := m.attendanceService.Mark(m.date, status, notes)
		if err != nil {
			return attendanceMarkErrorMsg(err.Error())
		}
		return attendanceMarkSuccessMsg{status: status}
	}
}

func getStatusIcon(status string) string {
	switch status {
	case "present":
		return "✓"
	case "remote":
		return "🏠"
	case "half-day":
		return "½"
	case "absent":
		return "✗"
	default:
		return "•"
	}
}

// Message types
type attendanceMarkSuccessMsg struct {
	status string
}
type attendanceMarkErrorMsg string

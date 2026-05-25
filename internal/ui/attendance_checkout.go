package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/attendance"
)

// AttendanceCheckOutModel represents the check-out form
type AttendanceCheckOutModel struct {
	attendanceService *attendance.Service
	notesInput        textinput.Model
	errorMsg          string
	successMsg        string
	loading           bool
	shouldGoBack      bool
}

// NewAttendanceCheckOutModel creates a new check-out model
func NewAttendanceCheckOutModel(attendanceService *attendance.Service) AttendanceCheckOutModel {
	notes := textinput.New()
	notes.Placeholder = "Optional notes..."
	notes.CharLimit = 500
	notes.Width = 50
	notes.Focus()

	return AttendanceCheckOutModel{
		attendanceService: attendanceService,
		notesInput:        notes,
	}
}

// Init initializes the check-out model
func (m AttendanceCheckOutModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m AttendanceCheckOutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			return m, m.handleSubmit()
		}

	case checkOutSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Checked out at %s", msg.time)
		return m, nil

	case checkOutErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update notes input
	var cmd tea.Cmd
	m.notesInput, cmd = m.notesInput.Update(msg)
	return m, cmd
}

// View renders the check-out form
func (m AttendanceCheckOutModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Check Out"))
	b.WriteString("\n\n")

	// Notes input
	b.WriteString(labelStyle.Render("Notes (optional):"))
	b.WriteString("\n")
	b.WriteString(focusedInputStyle.Render(m.notesInput.View()))
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
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Checking out..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		b.WriteString(helpStyle.Render("[Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

// handleSubmit handles the check-out form submission
func (m *AttendanceCheckOutModel) handleSubmit() tea.Cmd {
	notes := strings.TrimSpace(m.notesInput.Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		resp, err := m.attendanceService.CheckOut(notes)
		if err != nil {
			return checkOutErrorMsg(err.Error())
		}
		t := ""
		if resp.CheckoutTime != nil {
			t = utcTimeToLocal(*resp.CheckoutTime)
		}
		return checkOutSuccessMsg{time: t}
	}
}

// Message types
type checkOutSuccessMsg struct {
	time string
}
type checkOutErrorMsg string

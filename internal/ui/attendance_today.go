package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/attendance"
)

// AttendanceTodayModel represents the current user's today's attendance view
type AttendanceTodayModel struct {
	attendanceService *attendance.Service
	attendance        *api.AttendanceResponse
	errorMsg          string
	loading           bool
	shouldGoBack      bool
}

// NewAttendanceTodayModel creates a new attendance today model
func NewAttendanceTodayModel(attendanceService *attendance.Service) AttendanceTodayModel {
	return AttendanceTodayModel{
		attendanceService: attendanceService,
		loading:           true,
	}
}

// Init initializes the attendance today model
func (m AttendanceTodayModel) Init() tea.Cmd {
	return m.fetchAttendance()
}

// Update handles messages
func (m AttendanceTodayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, m.fetchAttendance()
		}

	case attendanceTodaySuccessMsg:
		m.loading = false
		m.attendance = msg.attendance
		return m, nil

	case attendanceTodayErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the today's attendance view
func (m AttendanceTodayModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Your Attendance Today"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading attendance..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if m.attendance == nil {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("You haven't marked attendance today."))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Use /checkin or /attendance to mark it."))
		b.WriteString("\n")
	} else {
		record := m.attendance
		timeStyle := lipgloss.NewStyle().Foreground(mutedColor)

		statusStyle := getStatusStyle(record.Status)
		b.WriteString(statusStyle.Render(fmt.Sprintf("%s %s", getStatusIcon(record.Status), record.Status)))
		b.WriteString("\n\n")

		inStr := "—"
		if record.CheckinAt != nil {
			inStr = utcTimeToLocal(*record.CheckinAt)
		}
		b.WriteString(timeStyle.Render(fmt.Sprintf("In:  %s", inStr)))
		b.WriteString("\n")

		outStr := "—"
		if record.CheckoutAt != nil {
			outStr = utcTimeToLocal(*record.CheckoutAt)
		}
		b.WriteString(timeStyle.Render(fmt.Sprintf("Out: %s", outStr)))
		b.WriteString("\n")

		if record.Notes != nil && *record.Notes != "" {
			notesStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
			b.WriteString("\n")
			b.WriteString(notesStyle.Render(*record.Notes))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("[r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

// fetchAttendance fetches the current user's attendance for today
func (m *AttendanceTodayModel) fetchAttendance() tea.Cmd {
	return func() tea.Msg {
		record, err := m.attendanceService.GetToday()
		if err != nil {
			return attendanceTodayErrorMsg(err.Error())
		}
		return attendanceTodaySuccessMsg{attendance: record}
	}
}

func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "present":
		return lipgloss.NewStyle().Foreground(successColor)
	case "remote":
		return lipgloss.NewStyle().Foreground(primaryColor)
	case "half-day":
		return lipgloss.NewStyle().Foreground(secondaryColor)
	case "absent":
		return lipgloss.NewStyle().Foreground(errorColor)
	default:
		return lipgloss.NewStyle().Foreground(mutedColor)
	}
}

// Message types
type attendanceTodaySuccessMsg struct {
	attendance *api.AttendanceResponse
}
type attendanceTodayErrorMsg string

package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/attendance"
)

// AttendanceHistoryModel represents the my attendance history view
type AttendanceHistoryModel struct {
	attendanceService *attendance.Service
	attendance        []*api.AttendanceResponse
	total             int
	errorMsg          string
	loading           bool
	shouldGoBack      bool
}

// NewAttendanceHistoryModel creates a new attendance history model
func NewAttendanceHistoryModel(attendanceService *attendance.Service) AttendanceHistoryModel {
	return AttendanceHistoryModel{
		attendanceService: attendanceService,
		loading:           true,
	}
}

// Init initializes the attendance history model
func (m AttendanceHistoryModel) Init() tea.Cmd {
	return m.fetchAttendance()
}

// Update handles messages
func (m AttendanceHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case attendanceHistorySuccessMsg:
		m.loading = false
		m.attendance = msg.attendance
		m.total = msg.total
		return m, nil

	case attendanceHistoryErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the my attendance history view
func (m AttendanceHistoryModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("My Attendance History"))
	b.WriteString("\n\n")

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading your attendance..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.attendance) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("You haven't marked any attendance yet"))
		b.WriteString("\n")
	} else {
		// Display attendance records
		for i, record := range m.attendance {
			if i > 0 {
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", 50)))
				b.WriteString("\n\n")
			}

			// Date
			dateStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			b.WriteString(dateStyle.Render(record.Date))
			b.WriteString("  ")

			// Status with icon
			statusStyle := getStatusStyle(record.Status)
			b.WriteString(statusStyle.Render(fmt.Sprintf("%s %s", getStatusIcon(record.Status), record.Status)))
			b.WriteString("\n")

			// Checkin/checkout times
			timeStyle := lipgloss.NewStyle().Foreground(mutedColor)
			if record.CheckinAt != nil {
				b.WriteString(timeStyle.Render(fmt.Sprintf("  In: %s", utcTimeToLocal(*record.CheckinAt))))
				if record.CheckoutAt != nil {
					b.WriteString(timeStyle.Render(fmt.Sprintf("  Out: %s", utcTimeToLocal(*record.CheckoutAt))))
				}
				b.WriteString("\n")
			}

			// Notes if present
			if record.Notes != nil && *record.Notes != "" {
				notesStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
				b.WriteString(notesStyle.Render("  " + *record.Notes))
				b.WriteString("\n")
			}
		}

		// Pagination info
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("Showing %d of %d records", len(m.attendance), m.total)))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	if !m.loading {
		b.WriteString(helpStyle.Render("[r] Refresh  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

// fetchAttendance fetches the user's attendance history
func (m *AttendanceHistoryModel) fetchAttendance() tea.Cmd {
	return func() tea.Msg {
		attendance, total, err := m.attendanceService.GetMy(10, 0)
		if err != nil {
			return attendanceHistoryErrorMsg(err.Error())
		}
		return attendanceHistorySuccessMsg{
			attendance: attendance,
			total:      total,
		}
	}
}

// Message types
type attendanceHistorySuccessMsg struct {
	attendance []*api.AttendanceResponse
	total      int
}
type attendanceHistoryErrorMsg string

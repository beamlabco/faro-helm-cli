package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/user"
)

// MemberOfficeHoursModel lets a primary admin set per-member office start time overrides
type MemberOfficeHoursModel struct {
	userService  *user.Service
	members      []*api.MemberResponse
	selectedUser int
	input        textinput.Model
	focusIndex   int // 0=member list, 1=time input
	errorMsg     string
	successMsg   string
	loading      bool
	shouldGoBack bool
}

// NewMemberOfficeHoursModel creates a new member office hours model
func NewMemberOfficeHoursModel(userService *user.Service) MemberOfficeHoursModel {
	input := textinput.New()
	input.Placeholder = "09:00  (blank to clear override)"
	input.CharLimit = 5
	input.Width = 30

	return MemberOfficeHoursModel{
		userService: userService,
		input:       input,
		focusIndex:  0,
		loading:     true,
	}
}

func (m MemberOfficeHoursModel) Init() tea.Cmd {
	return m.fetchMembers()
}

func (m MemberOfficeHoursModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "tab":
			if len(m.members) > 0 && !m.loading {
				if m.focusIndex == 0 {
					m.focusIndex = 1
					m.input.Focus()
					m.populateInputFromSelected()
				} else {
					m.focusIndex = 0
					m.input.Blur()
				}
			}
			return m, nil

		case "up", "k":
			if m.focusIndex == 0 && m.selectedUser > 0 {
				m.selectedUser--
				m.successMsg = ""
			}
			return m, nil

		case "down", "j":
			if m.focusIndex == 0 && m.selectedUser < len(m.members)-1 {
				m.selectedUser++
				m.successMsg = ""
			}
			return m, nil

		case "enter":
			if m.loading || len(m.members) == 0 {
				return m, nil
			}
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.input.Focus()
				m.populateInputFromSelected()
				return m, nil
			}
			return m, m.handleSubmit()
		}

	case memberOfficeHoursFetchMsg:
		m.loading = false
		m.members = msg.members
		if m.selectedUser >= len(m.members) {
			m.selectedUser = 0
		}
		return m, nil

	case memberOfficeHoursSuccessMsg:
		m.loading = false
		current := "cleared"
		if msg.officeStartTime != "" {
			current = msg.officeStartTime
		}
		m.successMsg = fmt.Sprintf("Updated %s's office start time: %s", msg.name, current)
		m.errorMsg = ""
		m.focusIndex = 0
		m.input.Blur()
		return m, m.fetchMembers()

	case memberOfficeHoursErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	if m.focusIndex == 1 {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *MemberOfficeHoursModel) populateInputFromSelected() {
	// Pre-fill with existing value if available (MemberResponse doesn't carry it,
	// so leave blank — the user can type a new value or leave blank to clear)
	m.input.SetValue("")
}

func (m MemberOfficeHoursModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Member Office Hours"))
	b.WriteString("\n\n")

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading team members..."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n\n")
	}

	if len(m.members) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No team members found."))
		b.WriteString("\n")
		return baseStyle.Render(b.String())
	}

	// Member list
	selectingLabel := ""
	if m.focusIndex == 0 {
		selectingLabel = " (selecting)"
	}
	b.WriteString(labelStyle.Render("Select member" + selectingLabel + ":"))
	b.WriteString("\n\n")

	for i, member := range m.members {
		prefix := "  "
		if i == m.selectedUser {
			prefix = "▸ "
		}

		var nameStyle lipgloss.Style
		if i == m.selectedUser {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
		} else {
			nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		}

		b.WriteString(prefix)
		b.WriteString(nameStyle.Render(member.Name))
		b.WriteString("  ")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(member.Email))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Time input
	b.WriteString(labelStyle.Render("Office start time (HH:MM):"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Leave blank to clear — member will use the org default."))
	b.WriteString("\n")
	if m.focusIndex == 1 {
		b.WriteString(focusedInputStyle.Render(m.input.View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.input.View()))
	}
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("[↑↓] Select member  [Tab/Enter] Next field  [Enter] Save  [Esc] Back"))

	return baseStyle.Render(b.String())
}

func (m *MemberOfficeHoursModel) fetchMembers() tea.Cmd {
	return func() tea.Msg {
		members, err := m.userService.GetMembers()
		if err != nil {
			return memberOfficeHoursErrorMsg(err.Error())
		}
		return memberOfficeHoursFetchMsg{members: members}
	}
}

func (m *MemberOfficeHoursModel) handleSubmit() tea.Cmd {
	if m.selectedUser >= len(m.members) {
		return nil
	}

	member := m.members[m.selectedUser]
	officeStartTime := strings.TrimSpace(m.input.Value())

	m.errorMsg = ""
	m.loading = true

	return func() tea.Msg {
		_, err := m.userService.UpdateOfficeHours(member.ID, officeStartTime)
		if err != nil {
			return memberOfficeHoursErrorMsg(err.Error())
		}
		return memberOfficeHoursSuccessMsg{name: member.Name, officeStartTime: officeStartTime}
	}
}

// Message types
type memberOfficeHoursFetchMsg struct {
	members []*api.MemberResponse
}
type memberOfficeHoursSuccessMsg struct {
	name            string
	officeStartTime string
}
type memberOfficeHoursErrorMsg string

package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/user"
)

// RoleUpdateModel represents the role update form with member selector
type RoleUpdateModel struct {
	userService   *user.Service
	members       []*api.MemberResponse
	selectedUser  int
	roleOptions   []string
	selectedRole  int
	focusIndex    int // 0=user, 1=role
	errorMsg      string
	successMsg    string
	loading       bool
	shouldGoBack  bool
}

// NewRoleUpdateModel creates a new role update model
func NewRoleUpdateModel(userService *user.Service) RoleUpdateModel {
	return RoleUpdateModel{
		userService: userService,
		roleOptions: []string{"manager", "member"},
		selectedRole: 0,
		focusIndex:   0,
		loading:      true,
	}
}

// Init initializes the model
func (m RoleUpdateModel) Init() tea.Cmd {
	return m.fetchMembers()
}

// Update handles messages
func (m RoleUpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "up", "k":
			if m.focusIndex == 0 && m.selectedUser > 0 {
				m.selectedUser--
			}
			return m, nil

		case "down", "j":
			if m.focusIndex == 0 && m.selectedUser < len(m.members)-1 {
				m.selectedUser++
			}
			return m, nil

		case "left", "h":
			if m.focusIndex == 1 {
				m.selectedRole--
				if m.selectedRole < 0 {
					m.selectedRole = len(m.roleOptions) - 1
				}
				return m, nil
			}

		case "right", "l":
			if m.focusIndex == 1 {
				m.selectedRole++
				if m.selectedRole >= len(m.roleOptions) {
					m.selectedRole = 0
				}
				return m, nil
			}

		case "tab":
			if len(m.members) > 0 {
				if m.focusIndex == 0 {
					m.focusIndex = 1
				} else {
					m.focusIndex = 0
				}
			}
			return m, nil

		case "enter":
			if m.loading || len(m.members) == 0 {
				return m, nil
			}
			return m, m.handleSubmit()

		case "r":
			if m.focusIndex == 0 {
				m.loading = true
				m.errorMsg = ""
				return m, m.fetchMembers()
			}
		}

	case roleUpdateFetchMsg:
		m.loading = false
		m.members = msg.members
		if m.selectedUser >= len(m.members) {
			m.selectedUser = 0
		}
		return m, nil

	case roleUpdateSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Updated %s's role to %s", msg.name, msg.role)
		// Refresh members to show updated roles
		return m, m.fetchMembers()

	case roleUpdateErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

// View renders the role update form
func (m RoleUpdateModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Update User Role"))
	b.WriteString("\n\n")

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading team members..."))
		b.WriteString("\n")
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	} else if len(m.members) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No team members found"))
		b.WriteString("\n")
	} else {
		// Member selector
		focusLabel := ""
		if m.focusIndex == 0 {
			focusLabel = " (selecting)"
		}
		b.WriteString(labelStyle.Render("Select member" + focusLabel + ":"))
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

			roleStyle := getRoleBadgeStyle(member.Role)
			emailStyle := lipgloss.NewStyle().Foreground(mutedColor)

			b.WriteString(prefix)
			b.WriteString(nameStyle.Render(member.Name))
			b.WriteString("  ")
			b.WriteString(roleStyle.Render(member.Role))
			b.WriteString("  ")
			b.WriteString(emailStyle.Render(member.Email))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// Role selection
		focusLabel = ""
		if m.focusIndex == 1 {
			focusLabel = " (selecting)"
		}
		b.WriteString(labelStyle.Render("New role" + focusLabel + ":"))
		b.WriteString("\n\n")

		for i, role := range m.roleOptions {
			var style lipgloss.Style
			if i == m.selectedRole {
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
			b.WriteString(style.Render(role))
		}
		b.WriteString("\n\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("[↑↓] Select member  [Tab] Switch field  [←→] Role  [Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *RoleUpdateModel) fetchMembers() tea.Cmd {
	return func() tea.Msg {
		members, err := m.userService.GetMembers()
		if err != nil {
			return roleUpdateErrorMsg(err.Error())
		}
		// Filter out primary users (their role can't be changed)
		var nonPrimary []*api.MemberResponse
		for _, member := range members {
			if member.Role != "primary" {
				nonPrimary = append(nonPrimary, member)
			}
		}
		return roleUpdateFetchMsg{members: nonPrimary}
	}
}

func (m *RoleUpdateModel) handleSubmit() tea.Cmd {
	if m.selectedUser >= len(m.members) {
		return nil
	}

	member := m.members[m.selectedUser]
	role := m.roleOptions[m.selectedRole]

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		resp, err := m.userService.UpdateRole(member.ID, role)
		if err != nil {
			return roleUpdateErrorMsg(err.Error())
		}
		return roleUpdateSuccessMsg{name: resp.Name, role: resp.Role}
	}
}

func getRoleBadgeStyle(role string) lipgloss.Style {
	switch role {
	case "primary":
		return lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	case "manager":
		return lipgloss.NewStyle().Foreground(secondaryColor)
	default:
		return lipgloss.NewStyle().Foreground(mutedColor)
	}
}

// Message types
type roleUpdateFetchMsg struct {
	members []*api.MemberResponse
}
type roleUpdateSuccessMsg struct {
	name string
	role string
}
type roleUpdateErrorMsg string

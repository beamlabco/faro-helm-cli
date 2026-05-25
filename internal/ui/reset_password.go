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

type resetPasswordPhase int

const (
	resetPasswordPhasePick resetPasswordPhase = iota
	resetPasswordPhaseSet
)

// ResetPasswordModel allows the primary user to reset a member's password
type ResetPasswordModel struct {
	userService  *user.Service
	members      []*api.MemberResponse
	selectedUser int
	phase        resetPasswordPhase
	input        textinput.Model
	confirmInput textinput.Model
	focusedInput int
	errorMsg     string
	successMsg   string
	loading      bool
	shouldGoBack bool
}

// NewResetPasswordModel creates a new reset password model
func NewResetPasswordModel(userService *user.Service) ResetPasswordModel {
	newPass := textinput.New()
	newPass.Placeholder = "New password (min 8 characters)"
	newPass.EchoMode = textinput.EchoPassword
	newPass.CharLimit = 255
	newPass.Width = 40

	confirm := textinput.New()
	confirm.Placeholder = "Confirm new password"
	confirm.EchoMode = textinput.EchoPassword
	confirm.CharLimit = 255
	confirm.Width = 40

	return ResetPasswordModel{
		userService:  userService,
		phase:        resetPasswordPhasePick,
		input:        newPass,
		confirmInput: confirm,
		loading:      true,
	}
}

// Init loads team members
func (m ResetPasswordModel) Init() tea.Cmd {
	return m.fetchMembers()
}

// Update handles messages
func (m ResetPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.phase == resetPasswordPhaseSet {
				// Go back to member picker
				m.phase = resetPasswordPhasePick
				m.errorMsg = ""
				m.successMsg = ""
				m.input.SetValue("")
				m.confirmInput.SetValue("")
				return m, nil
			}
			m.shouldGoBack = true
			return m, nil

		case "up", "k":
			if m.phase == resetPasswordPhasePick && m.selectedUser > 0 {
				m.selectedUser--
			}
			return m, nil

		case "down", "j":
			if m.phase == resetPasswordPhasePick && m.selectedUser < len(m.members)-1 {
				m.selectedUser++
			}
			return m, nil

		case "tab", "shift+tab":
			if m.phase == resetPasswordPhaseSet {
				if msg.String() == "tab" {
					m.focusedInput = 1 - m.focusedInput
				} else {
					m.focusedInput = 1 - m.focusedInput
				}
				if m.focusedInput == 0 {
					m.input.Focus()
					m.confirmInput.Blur()
				} else {
					m.input.Blur()
					m.confirmInput.Focus()
				}
			}
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			if m.phase == resetPasswordPhasePick {
				if len(m.members) == 0 {
					return m, nil
				}
				m.phase = resetPasswordPhaseSet
				m.focusedInput = 0
				m.input.Focus()
				m.confirmInput.Blur()
				m.errorMsg = ""
				m.successMsg = ""
				return m, textinput.Blink
			}
			return m, m.handleSubmit()

		case "r":
			if m.phase == resetPasswordPhasePick {
				m.loading = true
				m.errorMsg = ""
				return m, m.fetchMembers()
			}
		}

	case resetPasswordFetchMsg:
		m.loading = false
		m.members = msg.members
		if m.selectedUser >= len(m.members) {
			m.selectedUser = 0
		}
		return m, nil

	case resetPasswordSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Password reset for %s", msg.name)
		m.errorMsg = ""
		m.input.SetValue("")
		m.confirmInput.SetValue("")
		m.phase = resetPasswordPhasePick
		return m, nil

	case resetPasswordErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	if m.phase == resetPasswordPhaseSet {
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		m.confirmInput, cmd = m.confirmInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View renders the reset password UI
func (m ResetPasswordModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Reset User Password"))
	b.WriteString("\n\n")

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	if m.phase == resetPasswordPhasePick {
		if m.loading {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading team members..."))
			b.WriteString("\n")
		} else if len(m.members) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No team members found"))
			b.WriteString("\n")
		} else {
			b.WriteString(labelStyle.Render("Select member to reset password:"))
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

			if m.errorMsg != "" {
				b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
				b.WriteString("\n")
			}

			b.WriteString(helpStyle.Render("[↑↓] Select  [Enter] Next  [r] Refresh  [Esc] Back"))
		}
	} else {
		member := m.members[m.selectedUser]
		b.WriteString(labelStyle.Render(fmt.Sprintf("Setting new password for: %s (%s)", member.Name, member.Email)))
		b.WriteString("\n\n")

		b.WriteString(labelStyle.Render("New password:"))
		b.WriteString("\n")
		if m.focusedInput == 0 {
			b.WriteString(focusedInputStyle.Render(m.input.View()))
		} else {
			b.WriteString(blurredInputStyle.Render(m.input.View()))
		}
		b.WriteString("\n\n")

		b.WriteString(labelStyle.Render("Confirm new password:"))
		b.WriteString("\n")
		if m.focusedInput == 1 {
			b.WriteString(focusedInputStyle.Render(m.confirmInput.View()))
		} else {
			b.WriteString(blurredInputStyle.Render(m.confirmInput.View()))
		}
		b.WriteString("\n\n")

		if m.errorMsg != "" {
			b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
			b.WriteString("\n")
		}

		if m.loading {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Resetting password..."))
			b.WriteString("\n")
		}

		if !m.loading {
			b.WriteString(helpStyle.Render("[Tab] Switch field  [Enter] Confirm  [Esc] Back"))
		}
	}

	return baseStyle.Render(b.String())
}

func (m *ResetPasswordModel) fetchMembers() tea.Cmd {
	return func() tea.Msg {
		members, err := m.userService.GetMembers()
		if err != nil {
			return resetPasswordErrorMsg(err.Error())
		}
		// Exclude primary — their password can't be reset this way
		var nonPrimary []*api.MemberResponse
		for _, member := range members {
			if member.Role != "primary" {
				nonPrimary = append(nonPrimary, member)
			}
		}
		return resetPasswordFetchMsg{members: nonPrimary}
	}
}

func (m *ResetPasswordModel) handleSubmit() tea.Cmd {
	if m.selectedUser >= len(m.members) {
		return nil
	}

	newPassword := m.input.Value()
	confirmPassword := m.confirmInput.Value()

	m.errorMsg = ""

	if len(newPassword) < 8 {
		m.errorMsg = "Password must be at least 8 characters"
		return nil
	}
	if newPassword != confirmPassword {
		m.errorMsg = "Passwords do not match"
		return nil
	}

	member := m.members[m.selectedUser]
	m.loading = true

	return func() tea.Msg {
		if err := m.userService.ResetPassword(member.ID, newPassword); err != nil {
			return resetPasswordErrorMsg(err.Error())
		}
		return resetPasswordSuccessMsg{name: member.Name}
	}
}

// Message types
type resetPasswordFetchMsg struct {
	members []*api.MemberResponse
}
type resetPasswordSuccessMsg struct {
	name string
}
type resetPasswordErrorMsg string

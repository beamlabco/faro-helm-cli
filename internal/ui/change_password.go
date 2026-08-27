package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/user"
)

const (
	changePasswordCurrent = iota
	changePasswordNew
	changePasswordConfirm
)

// ChangePasswordModel represents the change password form
type ChangePasswordModel struct {
	userService  *user.Service
	inputs       []textinput.Model
	focusedInput int
	errorMsg     string
	successMsg   string
	loading      bool
	shouldGoBack bool
}

// NewChangePasswordModel creates a new change password model
func NewChangePasswordModel(userService *user.Service) ChangePasswordModel {
	inputs := make([]textinput.Model, 3)

	inputs[changePasswordCurrent] = textinput.New()
	inputs[changePasswordCurrent].Placeholder = "Current password"
	inputs[changePasswordCurrent].EchoMode = textinput.EchoPassword
	inputs[changePasswordCurrent].CharLimit = 255
	inputs[changePasswordCurrent].Width = 40

	inputs[changePasswordNew] = textinput.New()
	inputs[changePasswordNew].Placeholder = "New password (min 8 characters)"
	inputs[changePasswordNew].EchoMode = textinput.EchoPassword
	inputs[changePasswordNew].CharLimit = 255
	inputs[changePasswordNew].Width = 40

	inputs[changePasswordConfirm] = textinput.New()
	inputs[changePasswordConfirm].Placeholder = "Confirm new password"
	inputs[changePasswordConfirm].EchoMode = textinput.EchoPassword
	inputs[changePasswordConfirm].CharLimit = 255
	inputs[changePasswordConfirm].Width = 40

	return ChangePasswordModel{
		userService:  userService,
		inputs:       inputs,
		focusedInput: changePasswordCurrent,
	}
}

// Init initializes the model
func (m ChangePasswordModel) Init() tea.Cmd {
	m.inputs[changePasswordCurrent].Focus()
	return textinput.Blink
}

// Update handles messages
func (m ChangePasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			m.shouldGoBack = true
			return m, nil

		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusedInput--
			} else {
				m.focusedInput++
			}
			if m.focusedInput < 0 {
				m.focusedInput = len(m.inputs) - 1
			} else if m.focusedInput >= len(m.inputs) {
				m.focusedInput = 0
			}
			for i := range m.inputs {
				if i == m.focusedInput {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			return m, m.handleSubmit()
		}

	case changePasswordSuccessMsg:
		m.loading = false
		m.successMsg = "Password changed successfully!"
		m.errorMsg = ""
		m.inputs[changePasswordCurrent].SetValue("")
		m.inputs[changePasswordNew].SetValue("")
		m.inputs[changePasswordConfirm].SetValue("")
		return m, nil

	case changePasswordErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// View renders the change password form
func (m ChangePasswordModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Change Password"))
	b.WriteString("\n\n")

	labels := []string{"Current password:", "New password:", "Confirm new password:"}
	for i, label := range labels {
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")
		if m.focusedInput == i {
			b.WriteString(focusedInputStyle.Render(m.inputs[i].View()))
		} else {
			b.WriteString(blurredInputStyle.Render(m.inputs[i].View()))
		}
		b.WriteString("\n\n")
	}

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Changing password..."))
		b.WriteString("\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("[Tab] Next field  [Enter] Submit  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *ChangePasswordModel) handleSubmit() tea.Cmd {
	currentPassword := m.inputs[changePasswordCurrent].Value()
	newPassword := m.inputs[changePasswordNew].Value()
	confirmPassword := m.inputs[changePasswordConfirm].Value()

	m.errorMsg = ""
	m.successMsg = ""

	if currentPassword == "" {
		m.errorMsg = "Current password is required"
		return nil
	}
	if len(newPassword) < 8 {
		m.errorMsg = "New password must be at least 8 characters"
		return nil
	}
	if newPassword != confirmPassword {
		m.errorMsg = "Passwords do not match"
		return nil
	}

	m.loading = true
	return func() tea.Msg {
		if err := m.userService.ChangePassword(currentPassword, newPassword); err != nil {
			return changePasswordErrorMsg(err.Error())
		}
		return changePasswordSuccessMsg{}
	}
}

// Message types
type changePasswordSuccessMsg struct{}
type changePasswordErrorMsg string

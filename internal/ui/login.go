package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/auth"
)

// LoginModel represents the login screen
type LoginModel struct {
	authService   *auth.Service
	inputs        []textinput.Model
	focusedInput  int
	submitEnabled bool
	errorMsg      string
	successMsg    string
	loading       bool
	quitting      bool
	onSuccess     func()
}

const (
	loginEmailInput = iota
	loginPasswordInput
)

// NewLoginModel creates a new login model
func NewLoginModel(authService *auth.Service, onSuccess func()) LoginModel {
	m := LoginModel{
		authService:   authService,
		inputs:        make([]textinput.Model, 2),
		focusedInput:  loginEmailInput,
		submitEnabled: true,
		onSuccess:     onSuccess,
	}

	// Email input
	m.inputs[loginEmailInput] = textinput.New()
	m.inputs[loginEmailInput].Placeholder = "your@email.com"
	m.inputs[loginEmailInput].Focus()
	m.inputs[loginEmailInput].CharLimit = 255
	m.inputs[loginEmailInput].Width = 50

	// Password input
	m.inputs[loginPasswordInput] = textinput.New()
	m.inputs[loginPasswordInput].Placeholder = "••••••••"
	m.inputs[loginPasswordInput].EchoMode = textinput.EchoPassword
	m.inputs[loginPasswordInput].CharLimit = 255
	m.inputs[loginPasswordInput].Width = 50

	return m
}

// Init initializes the login model
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			m.quitting = true
			return m, nil

		case "tab", "shift+tab", "up", "down":
			// Cycle through inputs
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

			// Update focus
			for i := range m.inputs {
				if i == m.focusedInput {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, nil

		case "enter":
			if m.loading || !m.submitEnabled {
				return m, nil
			}
			return m, m.handleSubmit()
		}

	case loginSuccessMsg:
		m.loading = false
		m.successMsg = "Login successful!"
		if m.onSuccess != nil {
			m.onSuccess()
		}
		// Return message to shell - don't quit
		return m, nil

	case loginErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update inputs
	cmd := m.updateInputs(msg)
	return m, cmd
}

// View renders the login screen
func (m LoginModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🔐 Login to Faro Helm"))
	b.WriteString("\n\n")

	// Email input
	b.WriteString(labelStyle.Render("Email"))
	b.WriteString("\n")
	if m.focusedInput == loginEmailInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[loginEmailInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[loginEmailInput].View()))
	}
	b.WriteString("\n")

	// Password input
	b.WriteString(labelStyle.Render("Password"))
	b.WriteString("\n")
	if m.focusedInput == loginPasswordInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[loginPasswordInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[loginPasswordInput].View()))
	}
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("❌ " + m.errorMsg))
		b.WriteString("\n")
	}

	// Success message
	if m.successMsg != "" {
		b.WriteString(successStyle.Render("✓ " + m.successMsg))
		b.WriteString("\n")
	}

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("⏳ Logging in..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		b.WriteString(helpStyle.Render("tab: next • enter: submit • esc: quit"))
	}

	return baseStyle.Render(b.String())
}

// updateInputs updates the text inputs
func (m *LoginModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// handleSubmit handles the login form submission
func (m *LoginModel) handleSubmit() tea.Cmd {
	email := strings.TrimSpace(m.inputs[loginEmailInput].Value())
	password := m.inputs[loginPasswordInput].Value()

	// Clear previous messages
	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		if err := m.authService.Login(email, password); err != nil {
			return loginErrorMsg(err.Error())
		}
		return loginSuccessMsg{}
	}
}

// Message types
type loginSuccessMsg struct{}
type loginErrorMsg string

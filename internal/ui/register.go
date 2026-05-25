package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/auth"
)

// RegisterModel represents the registration screen
type RegisterModel struct {
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
	registerNameInput = iota
	registerEmailInput
	registerPasswordInput
	registerOrgNameInput
)

// NewRegisterModel creates a new register model
func NewRegisterModel(authService *auth.Service, onSuccess func()) RegisterModel {
	m := RegisterModel{
		authService:   authService,
		inputs:        make([]textinput.Model, 4),
		focusedInput:  registerNameInput,
		submitEnabled: true,
		onSuccess:     onSuccess,
	}

	// Name input
	m.inputs[registerNameInput] = textinput.New()
	m.inputs[registerNameInput].Placeholder = "John Doe"
	m.inputs[registerNameInput].Focus()
	m.inputs[registerNameInput].CharLimit = 255
	m.inputs[registerNameInput].Width = 50

	// Email input
	m.inputs[registerEmailInput] = textinput.New()
	m.inputs[registerEmailInput].Placeholder = "your@email.com"
	m.inputs[registerEmailInput].CharLimit = 255
	m.inputs[registerEmailInput].Width = 50

	// Password input
	m.inputs[registerPasswordInput] = textinput.New()
	m.inputs[registerPasswordInput].Placeholder = "••••••••"
	m.inputs[registerPasswordInput].EchoMode = textinput.EchoPassword
	m.inputs[registerPasswordInput].CharLimit = 255
	m.inputs[registerPasswordInput].Width = 50

	// Organization name input
	m.inputs[registerOrgNameInput] = textinput.New()
	m.inputs[registerOrgNameInput].Placeholder = "Acme Corp"
	m.inputs[registerOrgNameInput].CharLimit = 255
	m.inputs[registerOrgNameInput].Width = 50

	return m
}

// Init initializes the register model
func (m RegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case registerSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Welcome to Faro Helm, %s!", string(msg))
		if m.onSuccess != nil {
			m.onSuccess()
		}
		// Return message to shell - don't quit
		return m, nil

	case registerErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update inputs
	cmd := m.updateInputs(msg)
	return m, cmd
}

// View renders the registration screen
func (m RegisterModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📝 Register for Faro Helm"))
	b.WriteString("\n\n")

	// Name input
	b.WriteString(labelStyle.Render("Your Name"))
	b.WriteString("\n")
	if m.focusedInput == registerNameInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[registerNameInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[registerNameInput].View()))
	}
	b.WriteString("\n")

	// Email input
	b.WriteString(labelStyle.Render("Email"))
	b.WriteString("\n")
	if m.focusedInput == registerEmailInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[registerEmailInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[registerEmailInput].View()))
	}
	b.WriteString("\n")

	// Password input
	b.WriteString(labelStyle.Render("Password (minimum 8 characters)"))
	b.WriteString("\n")
	if m.focusedInput == registerPasswordInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[registerPasswordInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[registerPasswordInput].View()))
	}
	b.WriteString("\n")

	// Organization name input
	b.WriteString(labelStyle.Render("Organization Name"))
	b.WriteString("\n")
	if m.focusedInput == registerOrgNameInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[registerOrgNameInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[registerOrgNameInput].View()))
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
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("⏳ Creating your account..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		b.WriteString(helpStyle.Render("tab: next • enter: submit • esc: quit"))
	}

	return baseStyle.Render(b.String())
}

// updateInputs updates the text inputs
func (m *RegisterModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// handleSubmit handles the registration form submission
func (m *RegisterModel) handleSubmit() tea.Cmd {
	name := strings.TrimSpace(m.inputs[registerNameInput].Value())
	email := strings.TrimSpace(m.inputs[registerEmailInput].Value())
	password := m.inputs[registerPasswordInput].Value()
	orgName := strings.TrimSpace(m.inputs[registerOrgNameInput].Value())

	// Clear previous messages
	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		if err := m.authService.Register(email, password, name, orgName); err != nil {
			return registerErrorMsg(err.Error())
		}
		return registerSuccessMsg(name)
	}
}

// Message types
type registerSuccessMsg string
type registerErrorMsg string

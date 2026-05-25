package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/auth"
)

// JoinModel represents the join-via-invitation form
type JoinModel struct {
	authService  *auth.Service
	inputs       []textinput.Model
	focusedInput int
	errorMsg     string
	successMsg   string
	loading      bool
	quitting     bool
	onSuccess    func()
}

const (
	joinTokenInput = iota
	joinNameInput
	joinEmailInput
	joinPasswordInput
)

// NewJoinModel creates a new join model
func NewJoinModel(authService *auth.Service, onSuccess func()) JoinModel {
	m := JoinModel{
		authService:  authService,
		inputs:       make([]textinput.Model, 4),
		focusedInput: joinTokenInput,
		onSuccess:    onSuccess,
	}

	// Token input
	m.inputs[joinTokenInput] = textinput.New()
	m.inputs[joinTokenInput].Placeholder = "Paste your invitation token"
	m.inputs[joinTokenInput].Focus()
	m.inputs[joinTokenInput].CharLimit = 255
	m.inputs[joinTokenInput].Width = 50

	// Name input
	m.inputs[joinNameInput] = textinput.New()
	m.inputs[joinNameInput].Placeholder = "John Doe"
	m.inputs[joinNameInput].CharLimit = 255
	m.inputs[joinNameInput].Width = 50

	// Email input
	m.inputs[joinEmailInput] = textinput.New()
	m.inputs[joinEmailInput].Placeholder = "your@email.com"
	m.inputs[joinEmailInput].CharLimit = 255
	m.inputs[joinEmailInput].Width = 50

	// Password input
	m.inputs[joinPasswordInput] = textinput.New()
	m.inputs[joinPasswordInput].Placeholder = "••••••••"
	m.inputs[joinPasswordInput].EchoMode = textinput.EchoPassword
	m.inputs[joinPasswordInput].CharLimit = 255
	m.inputs[joinPasswordInput].Width = 50

	return m
}

// Init initializes the join model
func (m JoinModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m JoinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case joinSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Welcome to %s, %s!", msg.org, msg.name)
		if m.onSuccess != nil {
			m.onSuccess()
		}
		return m, nil

	case joinErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

// View renders the join form
func (m JoinModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Join an Organization"))
	b.WriteString("\n\n")

	labels := []string{"Invitation Token", "Your Name", "Email", "Password (minimum 8 characters)"}
	for i, label := range labels {
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")
		if m.focusedInput == i {
			b.WriteString(focusedInputStyle.Render(m.inputs[i].View()))
		} else {
			b.WriteString(blurredInputStyle.Render(m.inputs[i].View()))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Joining organization..."))
		b.WriteString("\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("tab: next • enter: submit • esc: back"))
	}

	return baseStyle.Render(b.String())
}

func (m *JoinModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m *JoinModel) handleSubmit() tea.Cmd {
	token := strings.TrimSpace(m.inputs[joinTokenInput].Value())
	name := strings.TrimSpace(m.inputs[joinNameInput].Value())
	email := strings.TrimSpace(m.inputs[joinEmailInput].Value())
	password := m.inputs[joinPasswordInput].Value()

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		if err := m.authService.Join(email, password, name, token); err != nil {
			return joinErrorMsg(err.Error())
		}
		org := m.authService.GetOrganization()
		orgName := ""
		if org != nil {
			orgName = org.Name
		}
		return joinSuccessMsg{name: name, org: orgName}
	}
}

// Message types
type joinSuccessMsg struct {
	name string
	org  string
}
type joinErrorMsg string

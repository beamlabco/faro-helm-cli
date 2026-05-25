package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/invitation"
)

// InvitationCreateModel represents the create invitation form
type InvitationCreateModel struct {
	invitationService *invitation.Service
	emailInput        textinput.Model
	errorMsg          string
	successMsg        string
	loading           bool
	shouldGoBack      bool
}

// NewInvitationCreateModel creates a new invitation create model
func NewInvitationCreateModel(invitationService *invitation.Service) InvitationCreateModel {
	email := textinput.New()
	email.Placeholder = "colleague@company.com"
	email.CharLimit = 255
	email.Width = 50
	email.Focus()

	return InvitationCreateModel{
		invitationService: invitationService,
		emailInput:        email,
	}
}

// Init initializes the model
func (m InvitationCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m InvitationCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case invitationCreateSuccessMsg:
		m.loading = false
		m.successMsg = fmt.Sprintf("Invitation sent to %s (token: %s)", msg.email, msg.token)
		return m, nil

	case invitationCreateErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	var cmd tea.Cmd
	m.emailInput, cmd = m.emailInput.Update(msg)
	return m, cmd
}

// View renders the form
func (m InvitationCreateModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Invite Team Member"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Email address:"))
	b.WriteString("\n")
	b.WriteString(focusedInputStyle.Render(m.emailInput.View()))
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Sending invitation..."))
		b.WriteString("\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("[Enter] Send  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *InvitationCreateModel) handleSubmit() tea.Cmd {
	email := strings.TrimSpace(m.emailInput.Value())

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		resp, err := m.invitationService.Create(email)
		if err != nil {
			return invitationCreateErrorMsg(err.Error())
		}
		return invitationCreateSuccessMsg{email: resp.Email, token: resp.Token}
	}
}

// Message types
type invitationCreateSuccessMsg struct {
	email string
	token string
}
type invitationCreateErrorMsg string

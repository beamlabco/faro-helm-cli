package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/project"
)

// ProjectCreateModel represents the project creation form
type ProjectCreateModel struct {
	projectService *project.Service
	nameInput      textinput.Model
	errorMsg       string
	successMsg     string
	loading        bool
	shouldGoBack   bool
}

// NewProjectCreateModel creates a new project creation model
func NewProjectCreateModel(projectService *project.Service) ProjectCreateModel {
	name := textinput.New()
	name.Placeholder = "My Project"
	name.Focus()
	name.CharLimit = 255
	name.Width = 40

	return ProjectCreateModel{
		projectService: projectService,
		nameInput:      name,
	}
}

// Init initializes the model
func (m ProjectCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m ProjectCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.shouldGoBack = true
			return m, nil
		case "enter":
			if !m.loading {
				return m, m.handleSubmit()
			}
		}

	case projectCreateSuccessMsg:
		m.loading = false
		m.successMsg = string(msg)
		return m, nil

	case projectCreateErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// View renders the form
func (m ProjectCreateModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Create Project"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Project name:"))
	b.WriteString("\n")
	b.WriteString(focusedInputStyle.Render(m.nameInput.View()))
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
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Creating project..."))
		b.WriteString("\n")
	}

	if !m.loading {
		b.WriteString(helpStyle.Render("[Enter] Create  [Esc] Back"))
	}

	return baseStyle.Render(b.String())
}

func (m *ProjectCreateModel) handleSubmit() tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())

	if name == "" {
		m.errorMsg = "Project name cannot be empty"
		return nil
	}

	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	return func() tea.Msg {
		p, err := m.projectService.Create(name)
		if err != nil {
			return projectCreateErrorMsg(err.Error())
		}
		return projectCreateSuccessMsg("Project \"" + p.Name + "\" created!")
	}
}

// Message types
type projectCreateSuccessMsg string
type projectCreateErrorMsg string

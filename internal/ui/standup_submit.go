package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/project"
	"github.com/beamlabco/faro-helm/internal/standup"
)

// StandupSubmitModel represents the standup submission form
type StandupSubmitModel struct {
	standupService *standup.Service
	projectService *project.Service
	inputs         []textarea.Model
	focusedInput   int
	date           string
	errorMsg       string
	successMsg     string
	loading        bool
	shouldGoBack   bool
	onBack         func()

	// Project selection
	projects       []*api.ProjectResponse
	selectedProject *api.ProjectResponse
	projectIdx     int
	phase          string // "project_select" or "form"
}

const (
	yesterdayInput = iota
	todayInput
	blockersInput
)

// NewStandupSubmitModel creates a new standup submission model
func NewStandupSubmitModel(standupService *standup.Service, projectService *project.Service, onBack func()) StandupSubmitModel {
	m := StandupSubmitModel{
		standupService: standupService,
		projectService: projectService,
		inputs:         make([]textarea.Model, 3),
		focusedInput:   yesterdayInput,
		date:           time.Now().Format("2006-01-02"),
		onBack:         onBack,
		phase:          "project_select",
		loading:        true,
	}

	// Yesterday textarea
	m.inputs[yesterdayInput] = textarea.New()
	m.inputs[yesterdayInput].Placeholder = "What did you work on yesterday?"
	m.inputs[yesterdayInput].CharLimit = 1000
	m.inputs[yesterdayInput].SetWidth(60)
	m.inputs[yesterdayInput].SetHeight(3)

	// Today textarea
	m.inputs[todayInput] = textarea.New()
	m.inputs[todayInput].Placeholder = "What will you work on today?"
	m.inputs[todayInput].CharLimit = 1000
	m.inputs[todayInput].SetWidth(60)
	m.inputs[todayInput].SetHeight(3)

	// Blockers textarea
	m.inputs[blockersInput] = textarea.New()
	m.inputs[blockersInput].Placeholder = "Any blockers or issues?"
	m.inputs[blockersInput].CharLimit = 1000
	m.inputs[blockersInput].SetWidth(60)
	m.inputs[blockersInput].SetHeight(3)

	return m
}

// Init initializes the standup submit model
func (m StandupSubmitModel) Init() tea.Cmd {
	return m.fetchProjects()
}

func (m *StandupSubmitModel) fetchProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectService.GetMyProjects()
		if err != nil {
			return standupProjectsErrorMsg(err.Error())
		}
		return standupProjectsLoadedMsg{projects: projects}
	}
}

// Update handles messages
func (m StandupSubmitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case standupProjectsLoadedMsg:
		m.loading = false
		m.projects = msg.projects
		if len(m.projects) == 1 {
			// Auto-select single project
			m.selectedProject = m.projects[0]
			m.phase = "form"
			m.inputs[yesterdayInput].Focus()
			return m, textarea.Blink
		}
		if len(m.projects) == 0 {
			m.errorMsg = "No projects found. Contact your admin."
			return m, nil
		}
		return m, nil

	case standupProjectsErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.phase == "form" && len(m.projects) > 1 {
				m.phase = "project_select"
				m.selectedProject = nil
				m.errorMsg = ""
				m.successMsg = ""
				return m, nil
			}
			m.shouldGoBack = true
			return m, nil

		case "up", "k":
			if m.phase == "project_select" {
				if m.projectIdx > 0 {
					m.projectIdx--
				}
				return m, nil
			}

		case "down", "j":
			if m.phase == "project_select" {
				if m.projectIdx < len(m.projects)-1 {
					m.projectIdx++
				}
				return m, nil
			}

		case "enter":
			if m.phase == "project_select" && len(m.projects) > 0 {
				m.selectedProject = m.projects[m.projectIdx]
				m.phase = "form"
				m.inputs[yesterdayInput].Focus()
				return m, textarea.Blink
			}

		case "tab", "shift+tab":
			if m.phase == "form" {
				if msg.String() == "shift+tab" {
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
			}

		case "ctrl+s":
			if m.phase == "form" && !m.loading {
				return m, m.handleSubmit()
			}
		}

	case standupSubmitSuccessMsg:
		m.loading = false
		m.successMsg = "Standup submitted successfully!"
		return m, nil

	case standupSubmitErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	// Update focused textarea
	if m.phase == "form" {
		var cmd tea.Cmd
		m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the standup submission form
func (m StandupSubmitModel) View() string {
	var b strings.Builder

	if m.loading {
		b.WriteString(titleStyle.Render("Submit Standup"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading projects..."))
		return baseStyle.Render(b.String())
	}

	if m.phase == "project_select" {
		b.WriteString(titleStyle.Render("Submit Standup - Select Project"))
		b.WriteString("\n\n")

		if m.errorMsg != "" {
			b.WriteString(errorStyle.Render(m.errorMsg))
			b.WriteString("\n\n")
			b.WriteString(helpStyle.Render("[esc] Back"))
			return baseStyle.Render(b.String())
		}

		for i, p := range m.projects {
			if i == m.projectIdx {
				b.WriteString(fmt.Sprintf("  > %s\n", lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(p.Name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", p.Name))
			}
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Navigate  [Enter] Select  [Esc] Back"))
		return baseStyle.Render(b.String())
	}

	// Form phase
	projectLabel := ""
	if m.selectedProject != nil {
		projectLabel = fmt.Sprintf(" [%s]", m.selectedProject.Name)
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("Submit Standup - %s%s", m.date, projectLabel)))
	b.WriteString("\n\n")

	// Yesterday input
	b.WriteString(labelStyle.Render("What did you do yesterday?"))
	b.WriteString("\n")
	if m.focusedInput == yesterdayInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[yesterdayInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[yesterdayInput].View()))
	}
	b.WriteString("\n")

	// Today input
	b.WriteString(labelStyle.Render("What will you do today?"))
	b.WriteString("\n")
	if m.focusedInput == todayInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[todayInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[todayInput].View()))
	}
	b.WriteString("\n")

	// Blockers input
	b.WriteString(labelStyle.Render("Any blockers?"))
	b.WriteString("\n")
	if m.focusedInput == blockersInput {
		b.WriteString(focusedInputStyle.Render(m.inputs[blockersInput].View()))
	} else {
		b.WriteString(blurredInputStyle.Render(m.inputs[blockersInput].View()))
	}
	b.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		b.WriteString("\n")
	}

	// Success message
	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n")
	}

	// Loading indicator
	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Submitting standup..."))
		b.WriteString("\n")
	}

	// Help text
	if !m.loading {
		helpText := "[Tab] Next field  [Ctrl+S] Submit  [Esc] Back"
		if len(m.projects) > 1 {
			helpText = "[Tab] Next field  [Ctrl+S] Submit  [Esc] Change project"
		}
		b.WriteString(helpStyle.Render(helpText))
	}

	return baseStyle.Render(b.String())
}

// handleSubmit handles the standup form submission
func (m *StandupSubmitModel) handleSubmit() tea.Cmd {
	yesterday := strings.TrimSpace(m.inputs[yesterdayInput].Value())
	today := strings.TrimSpace(m.inputs[todayInput].Value())
	blockers := strings.TrimSpace(m.inputs[blockersInput].Value())

	// Clear previous messages
	m.errorMsg = ""
	m.successMsg = ""
	m.loading = true

	projectID := m.selectedProject.ID

	return func() tea.Msg {
		if _, err := m.standupService.Submit(projectID, m.date, yesterday, today, blockers); err != nil {
			return standupSubmitErrorMsg(err.Error())
		}
		return standupSubmitSuccessMsg{}
	}
}

// Message types
type standupSubmitSuccessMsg struct{}
type standupSubmitErrorMsg string
type standupProjectsLoadedMsg struct {
	projects []*api.ProjectResponse
}
type standupProjectsErrorMsg string

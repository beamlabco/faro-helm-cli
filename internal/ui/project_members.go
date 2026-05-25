package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/project"
	"github.com/beamlabco/faro-helm/internal/user"
)

// ProjectMembersModel manages project membership
type ProjectMembersModel struct {
	projectService *project.Service
	userService    *user.Service

	// Data
	projects    []*api.ProjectResponse
	members     []*api.ProjectMemberResponse
	orgMembers  []*api.MemberResponse
	nonMembers  []*api.MemberResponse // org members not in project

	// UI state
	phase       string // "select", "view", "add"
	selectedIdx int
	cursorIdx   int
	selected    map[string]bool // for multi-select in add phase
	loading     bool
	errorMsg    string
	successMsg  string
	shouldGoBack bool
}

// NewProjectMembersModel creates a new project members model
func NewProjectMembersModel(projectService *project.Service, userService *user.Service) ProjectMembersModel {
	return ProjectMembersModel{
		projectService: projectService,
		userService:    userService,
		phase:          "select",
		loading:        true,
		selected:       make(map[string]bool),
	}
}

// Init loads projects
func (m ProjectMembersModel) Init() tea.Cmd {
	return m.fetchProjects()
}

// Update handles messages
func (m ProjectMembersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case "select":
			return m.updateSelectPhase(msg)
		case "view":
			return m.updateViewPhase(msg)
		case "add":
			return m.updateAddPhase(msg)
		}

	case projMembersProjectsMsg:
		m.loading = false
		m.projects = msg.projects
		if len(m.projects) == 0 {
			m.errorMsg = "No projects found"
		}
		return m, nil

	case projMembersMembersMsg:
		m.loading = false
		m.members = msg.members
		m.computeNonMembers()
		return m, nil

	case projMembersOrgMembersMsg:
		m.orgMembers = msg.members
		return m, m.fetchMembers(m.projects[m.selectedIdx].ID)

	case projMembersAddedMsg:
		m.loading = false
		m.successMsg = string(msg)
		m.selected = make(map[string]bool)
		m.cursorIdx = 0
		// Refresh members
		return m, m.fetchMembers(m.projects[m.selectedIdx].ID)

	case projMembersRemovedMsg:
		m.loading = false
		m.successMsg = string(msg)
		return m, m.fetchMembers(m.projects[m.selectedIdx].ID)

	case projMembersErrorMsg:
		m.loading = false
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, nil
}

func (m ProjectMembersModel) updateSelectPhase(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.shouldGoBack = true
		return m, nil
	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
	case "down", "j":
		if m.selectedIdx < len(m.projects)-1 {
			m.selectedIdx++
		}
	case "enter":
		if len(m.projects) > 0 {
			m.phase = "view"
			m.loading = true
			m.errorMsg = ""
			m.successMsg = ""
			m.cursorIdx = 0
			return m, m.fetchOrgAndProjectMembers()
		}
	}
	return m, nil
}

func (m ProjectMembersModel) updateViewPhase(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.phase = "select"
		m.errorMsg = ""
		m.successMsg = ""
		return m, nil
	case "up", "k":
		if m.cursorIdx > 0 {
			m.cursorIdx--
		}
	case "down", "j":
		if m.cursorIdx < len(m.members)-1 {
			m.cursorIdx++
		}
	case "a":
		if len(m.nonMembers) > 0 {
			m.phase = "add"
			m.cursorIdx = 0
			m.selected = make(map[string]bool)
			m.errorMsg = ""
			m.successMsg = ""
		}
	case "d":
		if len(m.members) > 0 && !m.loading {
			member := m.members[m.cursorIdx]
			m.loading = true
			m.errorMsg = ""
			m.successMsg = ""
			return m, m.removeMember(m.projects[m.selectedIdx].ID, member.ID, member.Name)
		}
	}
	return m, nil
}

func (m ProjectMembersModel) updateAddPhase(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.phase = "view"
		m.errorMsg = ""
		m.successMsg = ""
		return m, nil
	case "up", "k":
		if m.cursorIdx > 0 {
			m.cursorIdx--
		}
	case "down", "j":
		if m.cursorIdx < len(m.nonMembers)-1 {
			m.cursorIdx++
		}
	case " ":
		if len(m.nonMembers) > 0 {
			userID := m.nonMembers[m.cursorIdx].ID
			m.selected[userID] = !m.selected[userID]
			if !m.selected[userID] {
				delete(m.selected, userID)
			}
		}
	case "enter":
		if len(m.selected) > 0 && !m.loading {
			m.loading = true
			m.errorMsg = ""
			m.successMsg = ""
			return m, m.addMembers(m.projects[m.selectedIdx].ID)
		}
	}
	return m, nil
}

// View renders the model
func (m ProjectMembersModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Project Members"))
	b.WriteString("\n\n")

	if m.loading && m.phase == "select" {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading projects..."))
		return baseStyle.Render(b.String())
	}

	switch m.phase {
	case "select":
		b.WriteString(labelStyle.Render("Select a project:"))
		b.WriteString("\n\n")

		for i, p := range m.projects {
			if i == m.selectedIdx {
				b.WriteString(fmt.Sprintf("  > %s\n", lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(p.Name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s\n", p.Name))
			}
		}

		if m.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Navigate  [Enter] Select  [Esc] Back"))

	case "view":
		proj := m.projects[m.selectedIdx]
		b.WriteString(labelStyle.Render(fmt.Sprintf("Members of %s:", proj.Name)))
		b.WriteString("\n\n")

		if m.loading {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading members..."))
			b.WriteString("\n")
		} else if len(m.members) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("No members in this project."))
			b.WriteString("\n")
		} else {
			nameStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
			dimStyle := lipgloss.NewStyle().Foreground(mutedColor)

			for i, member := range m.members {
				cursor := "  "
				if i == m.cursorIdx {
					cursor = "> "
				}
				roleTag := ""
				if member.Role == "primary" {
					roleTag = dimStyle.Render(" (admin)")
				}
				b.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, nameStyle.Render(member.Name), dimStyle.Render(member.Email), roleTag))
			}
		}

		if m.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		}
		if m.successMsg != "" {
			b.WriteString("\n")
			b.WriteString(successStyle.Render(m.successMsg))
		}

		b.WriteString("\n")
		addHint := ""
		if len(m.nonMembers) > 0 {
			addHint = "[a] Add members  "
		}
		removeHint := ""
		if len(m.members) > 0 {
			removeHint = "[d] Remove selected  "
		}
		b.WriteString(helpStyle.Render(fmt.Sprintf("[↑↓] Navigate  %s%s[Esc] Back", addHint, removeHint)))

	case "add":
		proj := m.projects[m.selectedIdx]
		b.WriteString(labelStyle.Render(fmt.Sprintf("Add members to %s:", proj.Name)))
		b.WriteString("\n\n")

		if m.loading {
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("Adding members..."))
			b.WriteString("\n")
		} else {
			nameStyle := lipgloss.NewStyle().Bold(true)
			dimStyle := lipgloss.NewStyle().Foreground(mutedColor)
			selectedStyle := lipgloss.NewStyle().Foreground(successColor)

			for i, member := range m.nonMembers {
				cursor := "  "
				if i == m.cursorIdx {
					cursor = "> "
				}
				checkbox := "[ ]"
				nameColor := nameStyle
				if m.selected[member.ID] {
					checkbox = "[x]"
					nameColor = nameStyle.Foreground(successColor)
				}
				b.WriteString(fmt.Sprintf("  %s%s %s %s\n", cursor, selectedStyle.Render(checkbox), nameColor.Render(member.Name), dimStyle.Render(member.Email)))
			}

			selectedCount := len(m.selected)
			if selectedCount > 0 {
				b.WriteString(fmt.Sprintf("\n  %s selected", lipgloss.NewStyle().Bold(true).Foreground(secondaryColor).Render(fmt.Sprintf("%d", selectedCount))))
			}
		}

		if m.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
		}

		b.WriteString("\n")
		enterHint := ""
		if len(m.selected) > 0 {
			enterHint = "[Enter] Confirm  "
		}
		b.WriteString(helpStyle.Render(fmt.Sprintf("[↑↓] Navigate  [Space] Toggle  %s[Esc] Back", enterHint)))
	}

	return baseStyle.Render(b.String())
}

func (m *ProjectMembersModel) computeNonMembers() {
	memberIDs := make(map[string]bool)
	for _, mem := range m.members {
		memberIDs[mem.ID] = true
	}

	m.nonMembers = nil
	for _, om := range m.orgMembers {
		if !memberIDs[om.ID] {
			m.nonMembers = append(m.nonMembers, om)
		}
	}
}

// Commands

func (m *ProjectMembersModel) fetchProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectService.GetAll()
		if err != nil {
			return projMembersErrorMsg(err.Error())
		}
		return projMembersProjectsMsg{projects: projects}
	}
}

func (m *ProjectMembersModel) fetchOrgAndProjectMembers() tea.Cmd {
	return func() tea.Msg {
		members, err := m.userService.GetMembers()
		if err != nil {
			return projMembersErrorMsg(err.Error())
		}
		return projMembersOrgMembersMsg{members: members}
	}
}

func (m *ProjectMembersModel) fetchMembers(projectID string) tea.Cmd {
	return func() tea.Msg {
		members, err := m.projectService.GetMembers(projectID)
		if err != nil {
			return projMembersErrorMsg(err.Error())
		}
		return projMembersMembersMsg{members: members}
	}
}

func (m *ProjectMembersModel) addMembers(projectID string) tea.Cmd {
	userIDs := make([]string, 0, len(m.selected))
	for id := range m.selected {
		userIDs = append(userIDs, id)
	}

	return func() tea.Msg {
		err := m.projectService.AddMembers(projectID, userIDs)
		if err != nil {
			return projMembersErrorMsg(err.Error())
		}
		count := len(userIDs)
		if count == 1 {
			return projMembersAddedMsg("1 member added!")
		}
		return projMembersAddedMsg(fmt.Sprintf("%d members added!", count))
	}
}

func (m *ProjectMembersModel) removeMember(projectID, userID string, name string) tea.Cmd {
	return func() tea.Msg {
		err := m.projectService.RemoveMember(projectID, userID)
		if err != nil {
			return projMembersErrorMsg(err.Error())
		}
		return projMembersRemovedMsg(fmt.Sprintf("%s removed from project", name))
	}
}

// Message types
type projMembersProjectsMsg struct {
	projects []*api.ProjectResponse
}
type projMembersMembersMsg struct {
	members []*api.ProjectMemberResponse
}
type projMembersOrgMembersMsg struct {
	members []*api.MemberResponse
}
type projMembersAddedMsg string
type projMembersRemovedMsg string
type projMembersErrorMsg string

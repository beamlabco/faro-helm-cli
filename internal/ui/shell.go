package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/attendance"
	"github.com/beamlabco/faro-helm-cli/internal/auth"
	"github.com/beamlabco/faro-helm-cli/internal/config"
	"github.com/beamlabco/faro-helm-cli/internal/leave"
	"github.com/beamlabco/faro-helm-cli/internal/team"
	"github.com/beamlabco/faro-helm-cli/internal/standup"
	"github.com/beamlabco/faro-helm-cli/internal/user"
)

// View states
type viewState int

const (
	viewShell viewState = iota
	viewDeviceFlow
	viewLogin
	viewRegister
	viewStandupSubmit
	viewStandupToday
	viewStandupHistory
	viewAttendanceCheckIn
	viewAttendanceCheckOut
	viewAttendanceToday
	viewAttendanceHistory
	viewJoin
	viewLeaveRequest
	viewLeaveList
	viewLeaveCancel
	viewPeopleList
	viewTeamList
	viewChangePassword
)

// Dashboard message types
type dashboardStandupMsg struct {
	standups []*api.StandupResponse
	err      error
}

type dashboardAttendanceMsg struct {
	resp *api.AttendanceResponse
	err  error
}

type dashboardApprovedLeavesMsg struct {
	onLeave []string
	err     error
}

// ShellModel is the main REPL shell
type ShellModel struct {
	// Services
	authService       *auth.Service
	standupService    *standup.Service
	attendanceService *attendance.Service
	leaveService      *leave.Service
	userService       *user.Service
	teamService    *team.Service
	config            *config.Config

	// Shell components
	commandInput *CommandInput
	registry     *CommandRegistry

	// Current view state
	currentView viewState

	// Sub-models
	deviceFlowModel        DeviceFlowModel
	loginModel             LoginModel
	registerModel          RegisterModel
	standupSubmitModel     StandupSubmitModel
	standupTodayModel      StandupTodayModel
	standupHistoryModel    StandupHistoryModel
	attendanceCheckInModel   AttendanceCheckInModel
	attendanceCheckOutModel  AttendanceCheckOutModel
	attendanceTodayModel     AttendanceTodayModel
	attendanceHistoryModel   AttendanceHistoryModel
	joinModel                JoinModel
	leaveRequestModel        LeaveRequestModel
	leaveListModel           LeaveListModel
	leaveCancelModel         LeaveCancelModel
	peopleListModel            PeopleListModel
	teamListModel         TeamListModel
	changePasswordModel      ChangePasswordModel

	// Output area
	output     string
	outputType string // "success", "error", "info"

	// Dashboard state
	dashboard struct {
		loading          bool
		myStandup        *api.StandupResponse
		myAttendance     *api.AttendanceResponse
		onLeaveToday     []string
		fetchesCompleted int
		totalFetches     int
	}

	// State
	quitting bool
	width    int
	height   int
}

// NewShellModel creates a new shell model
func NewShellModel(authService *auth.Service, standupService *standup.Service, attendanceService *attendance.Service, leaveService *leave.Service, userService *user.Service, teamService *team.Service, cfg *config.Config) ShellModel {
	registry := NewCommandRegistry()
	isAuth := authService.IsAuthenticated()
	role := ""
	if isAuth {
		if u := authService.GetUser(); u != nil {
			role = u.Role
		}
	}

	return ShellModel{
		authService:       authService,
		standupService:    standupService,
		attendanceService: attendanceService,
		leaveService:      leaveService,
		userService:       userService,
		teamService:    teamService,
		config:            cfg,
		commandInput:      NewCommandInput(registry, isAuth, role),
		registry:          registry,
		currentView:       viewShell,
		width:             80,
		height:            24,
	}
}

// Init initializes the shell
func (m ShellModel) Init() tea.Cmd {
	if m.authService.IsAuthenticated() {
		m.dashboard.loading = true
		m.dashboard.totalFetches = 3
		return tea.Batch(
			m.commandInput.Focus(),
			m.fetchDashboardStandups(),
			m.fetchDashboardAttendance(),
			m.fetchDashboardApprovedLeaves(),
		)
	}
	return m.commandInput.Focus()
}

func (m ShellModel) fetchDashboardStandups() tea.Cmd {
	return func() tea.Msg {
		standups, err := m.standupService.GetToday()
		return dashboardStandupMsg{standups: standups, err: err}
	}
}

func (m ShellModel) fetchDashboardAttendance() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.attendanceService.GetToday()
		return dashboardAttendanceMsg{resp: resp, err: err}
	}
}

func (m ShellModel) fetchDashboardApprovedLeaves() tea.Cmd {
	return func() tea.Msg {
		today := time.Now().Format("2006-01-02")
		leaves, _, err := m.leaveService.GetAll("approved", "", 100, 0)
		if err != nil {
			return dashboardApprovedLeavesMsg{err: err}
		}
		var onLeave []string
		for _, l := range leaves {
			if l.StartDate <= today && l.EndDate >= today && l.User != nil {
				onLeave = append(onLeave, l.User.Name)
			}
		}
		return dashboardApprovedLeavesMsg{onLeave: onLeave}
	}
}

func (m *ShellModel) startDashboardFetch() tea.Cmd {
	m.dashboard = struct {
		loading          bool
		myStandup        *api.StandupResponse
		myAttendance     *api.AttendanceResponse
		onLeaveToday     []string
		fetchesCompleted int
		totalFetches     int
	}{loading: true, totalFetches: 3}
	return tea.Batch(
		m.fetchDashboardStandups(),
		m.fetchDashboardAttendance(),
		m.fetchDashboardApprovedLeaves(),
	)
}

// Update handles messages
func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle dashboard messages
	switch msg := msg.(type) {
	case dashboardStandupMsg:
		m.dashboard.fetchesCompleted++
		if msg.err == nil && len(msg.standups) > 0 {
			m.dashboard.myStandup = msg.standups[0]
		}
		if m.dashboard.fetchesCompleted >= m.dashboard.totalFetches {
			m.dashboard.loading = false
		}
		return m, nil
	case dashboardAttendanceMsg:
		m.dashboard.fetchesCompleted++
		if msg.err == nil {
			m.dashboard.myAttendance = msg.resp
		}
		if m.dashboard.fetchesCompleted >= m.dashboard.totalFetches {
			m.dashboard.loading = false
		}
		return m, nil
	case dashboardApprovedLeavesMsg:
		m.dashboard.fetchesCompleted++
		if msg.err == nil {
			m.dashboard.onLeaveToday = msg.onLeave
		}
		if m.dashboard.fetchesCompleted >= m.dashboard.totalFetches {
			m.dashboard.loading = false
		}
		return m, nil
	}

	// Route to sub-model if active
	switch m.currentView {
	case viewDeviceFlow:
		return m.updateDeviceFlow(msg)
	case viewLogin:
		return m.updateLogin(msg)
	case viewRegister:
		return m.updateRegister(msg)
	case viewStandupSubmit:
		return m.updateStandupSubmit(msg)
	case viewStandupToday:
		return m.updateStandupToday(msg)
	case viewStandupHistory:
		return m.updateStandupHistory(msg)
	case viewAttendanceCheckIn:
		return m.updateAttendanceCheckIn(msg)
	case viewAttendanceCheckOut:
		return m.updateAttendanceCheckOut(msg)
	case viewAttendanceToday:
		return m.updateAttendanceToday(msg)
	case viewAttendanceHistory:
		return m.updateAttendanceHistory(msg)
	case viewJoin:
		return m.updateJoin(msg)
	case viewLeaveRequest:
		return m.updateLeaveRequest(msg)
	case viewLeaveList:
		return m.updateLeaveList(msg)
	case viewLeaveCancel:
		return m.updateLeaveCancel(msg)
	case viewPeopleList:
		return m.updatePeopleList(msg)
	case viewTeamList:
		return m.updateTeamList(msg)
	case viewChangePassword:
		return m.updateChangePassword(msg)
	}

	// Handle shell input
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+l":
			// Clear screen
			m.output = ""
			m.outputType = ""
			return m, nil

		case "enter":
			// If dropdown is visible and has a selection, execute that command directly
			if selected := m.commandInput.SelectedCommand(); selected != nil {
				m.commandInput.Reset()
				// Check auth requirement
				if selected.RequiresAuth && !m.authService.IsAuthenticated() {
					m.output = fmt.Sprintf("You must be logged in to use /%s. Use /login first.", selected.Name)
					m.outputType = "error"
					return m, m.commandInput.Focus()
				}
				return m.handleCommand(selected.Name)
			}
			// Otherwise execute the typed command
			return m.executeCommand()
		}
	}

	// Update command input
	var cmd tea.Cmd
	m.commandInput, cmd = m.commandInput.Update(msg)
	return m, cmd
}

// executeCommand executes the current command
func (m ShellModel) executeCommand() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.commandInput.Value())
	m.commandInput.Reset()

	if input == "" {
		return m, m.commandInput.Focus()
	}

	// Must start with /
	if !strings.HasPrefix(input, "/") {
		m.output = "Commands must start with /"
		m.outputType = "error"
		return m, m.commandInput.Focus()
	}

	// Check for empty command (just "/")
	if input == "/" {
		m.output = "Type a command after /. Use /help for available commands."
		m.outputType = "error"
		return m, m.commandInput.Focus()
	}

	// Find command
	cmd := m.registry.FindExact(input)
	if cmd == nil {
		m.output = fmt.Sprintf("Unknown command: %s. Type /help for available commands.", input)
		m.outputType = "error"
		return m, m.commandInput.Focus()
	}

	// Check auth requirement
	if cmd.RequiresAuth && !m.authService.IsAuthenticated() {
		m.output = fmt.Sprintf("You must be logged in to use /%s. Use /login first.", cmd.Name)
		m.outputType = "error"
		return m, m.commandInput.Focus()
	}

	// Execute command
	return m.handleCommand(cmd.Name)
}

// handleCommand handles command execution
func (m ShellModel) handleCommand(cmdName string) (tea.Model, tea.Cmd) {
	switch cmdName {
	// Auth commands
	case "login":
		m.currentView = viewDeviceFlow
		m.deviceFlowModel = NewDeviceFlowModel(m.authService, nil)
		return m, m.deviceFlowModel.Init()

	case "signup":
		m.currentView = viewRegister
		m.registerModel = NewRegisterModel(m.authService, nil)
		return m, m.registerModel.Init()

	case "join":
		m.currentView = viewJoin
		m.joinModel = NewJoinModel(m.authService, nil)
		return m, m.joinModel.Init()

	case "logout":
		if err := m.authService.Logout(); err != nil {
			m.output = fmt.Sprintf("Logout failed: %s", err.Error())
			m.outputType = "error"
		} else {
			m.output = ""
			m.outputType = ""
			m.commandInput.SetAuth(false, "")
			m.dashboard = struct {
				loading          bool
				myStandup        *api.StandupResponse
				myAttendance     *api.AttendanceResponse
				onLeaveToday     []string
				fetchesCompleted int
				totalFetches     int
			}{}
		}
		return m, m.commandInput.Focus()

	// Team commands
	case "team":
		m.currentView = viewTeamList
		m.teamListModel = NewTeamListModel(m.teamService)
		return m, m.teamListModel.Init()

	// Standup commands
	case "standup":
		m.currentView = viewStandupSubmit
		m.standupSubmitModel = NewStandupSubmitModel(m.standupService, m.teamService, nil)
		return m, m.standupSubmitModel.Init()

	case "standup today":
		m.currentView = viewStandupToday
		m.standupTodayModel = NewStandupTodayModel(m.standupService, nil)
		return m, m.standupTodayModel.Init()

	case "standup history":
		m.currentView = viewStandupHistory
		m.standupHistoryModel = NewStandupHistoryModel(m.standupService, nil)
		return m, m.standupHistoryModel.Init()

	// Attendance commands
	case "checkin":
		m.currentView = viewAttendanceCheckIn
		m.attendanceCheckInModel = NewAttendanceCheckInModel(m.attendanceService)
		return m, m.attendanceCheckInModel.Init()

	case "checkout":
		m.currentView = viewAttendanceCheckOut
		m.attendanceCheckOutModel = NewAttendanceCheckOutModel(m.attendanceService)
		return m, m.attendanceCheckOutModel.Init()

	case "attendance today":
		m.currentView = viewAttendanceToday
		m.attendanceTodayModel = NewAttendanceTodayModel(m.attendanceService)
		return m, m.attendanceTodayModel.Init()

	case "attendance history":
		m.currentView = viewAttendanceHistory
		m.attendanceHistoryModel = NewAttendanceHistoryModel(m.attendanceService)
		return m, m.attendanceHistoryModel.Init()

	// Leave commands
	case "leave":
		m.currentView = viewLeaveRequest
		m.leaveRequestModel = NewLeaveRequestModel(m.leaveService)
		return m, m.leaveRequestModel.Init()

	case "leave list":
		m.currentView = viewLeaveList
		m.leaveListModel = NewLeaveListModel(m.leaveService)
		return m, m.leaveListModel.Init()

	case "leave cancel":
		userID := ""
		if m.config.User != nil {
			userID = m.config.User.ID
		}
		m.currentView = viewLeaveCancel
		m.leaveCancelModel = NewLeaveCancelModel(m.leaveService, userID)
		return m, m.leaveCancelModel.Init()

	// People
	case "people":
		m.currentView = viewPeopleList
		m.peopleListModel = NewPeopleListModel(m.userService)
		return m, m.peopleListModel.Init()

	case "password":
		m.currentView = viewChangePassword
		m.changePasswordModel = NewChangePasswordModel(m.userService)
		return m, m.changePasswordModel.Init()

	// Utility commands
	case "help":
		m.output = m.renderHelp()
		m.outputType = "info"
		return m, m.commandInput.Focus()

	case "whoami":
		m.output = m.renderWhoami()
		m.outputType = "info"
		return m, m.commandInput.Focus()

	case "clear":
		m.output = ""
		m.outputType = ""
		return m, m.commandInput.Focus()

	case "quit":
		m.quitting = true
		return m, tea.Quit

	default:
		m.output = fmt.Sprintf("Command not implemented: %s", cmdName)
		m.outputType = "error"
		return m, m.commandInput.Focus()
	}
}

// Sub-model update handlers

func (m ShellModel) updateDeviceFlow(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case deviceFlowSuccessMsg:
		m.currentView = viewShell
		m.output = ""
		m.outputType = ""
		m.commandInput.SetAuth(true, m.getUserRole())
		cmd := m.startDashboardFetch()
		return m, tea.Batch(m.commandInput.Focus(), cmd)
	}

	newModel, cmd := m.deviceFlowModel.Update(msg)
	m.deviceFlowModel = newModel.(DeviceFlowModel)

	if m.deviceFlowModel.quitting {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for escape or success
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case loginSuccessMsg:
		m.currentView = viewShell
		m.output = ""
		m.outputType = ""
		m.commandInput.SetAuth(true, m.getUserRole())
		cmd := m.startDashboardFetch()
		return m, tea.Batch(m.commandInput.Focus(), cmd)
	case loginErrorMsg:
		// Let the login model handle displaying the error
	}

	// Update login model
	newModel, cmd := m.loginModel.Update(msg)
	m.loginModel = newModel.(LoginModel)

	// Check if quitting (user pressed ctrl+c)
	if m.loginModel.quitting {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateRegister(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case registerSuccessMsg:
		m.currentView = viewShell
		m.output = ""
		m.outputType = ""
		m.commandInput.SetAuth(true, m.getUserRole())
		cmd := m.startDashboardFetch()
		return m, tea.Batch(m.commandInput.Focus(), cmd)
	}

	newModel, cmd := m.registerModel.Update(msg)
	m.registerModel = newModel.(RegisterModel)

	if m.registerModel.quitting {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateStandupSubmit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case standupSubmitSuccessMsg:
		m.currentView = viewShell
		m.output = "Standup submitted successfully!"
		m.outputType = "success"
		return m, m.commandInput.Focus()
	}

	newModel, cmd := m.standupSubmitModel.Update(msg)
	m.standupSubmitModel = newModel.(StandupSubmitModel)

	if m.standupSubmitModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateStandupToday(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.standupTodayModel.Update(msg)
	m.standupTodayModel = newModel.(StandupTodayModel)

	if m.standupTodayModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateStandupHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.standupHistoryModel.Update(msg)
	m.standupHistoryModel = newModel.(StandupHistoryModel)

	if m.standupHistoryModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateAttendanceCheckIn(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case checkInSuccessMsg:
		m.currentView = viewShell
		m.output = fmt.Sprintf("Checked in as %s at %s", msg.status, msg.time)
		m.outputType = "success"
		return m, m.commandInput.Focus()
	}

	newModel, cmd := m.attendanceCheckInModel.Update(msg)
	m.attendanceCheckInModel = newModel.(AttendanceCheckInModel)

	if m.attendanceCheckInModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateAttendanceCheckOut(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case checkOutSuccessMsg:
		m.currentView = viewShell
		m.output = fmt.Sprintf("Checked out at %s", msg.time)
		m.outputType = "success"
		return m, m.commandInput.Focus()
	}

	newModel, cmd := m.attendanceCheckOutModel.Update(msg)
	m.attendanceCheckOutModel = newModel.(AttendanceCheckOutModel)

	if m.attendanceCheckOutModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateAttendanceToday(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.attendanceTodayModel.Update(msg)
	m.attendanceTodayModel = newModel.(AttendanceTodayModel)

	if m.attendanceTodayModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateAttendanceHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.attendanceHistoryModel.Update(msg)
	m.attendanceHistoryModel = newModel.(AttendanceHistoryModel)

	if m.attendanceHistoryModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateJoin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case joinSuccessMsg:
		m.currentView = viewShell
		m.output = ""
		m.outputType = ""
		m.commandInput.SetAuth(true, m.getUserRole())
		cmd := m.startDashboardFetch()
		return m, tea.Batch(m.commandInput.Focus(), cmd)
	}

	newModel, cmd := m.joinModel.Update(msg)
	m.joinModel = newModel.(JoinModel)

	if m.joinModel.quitting {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateLeaveRequest(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	case leaveRequestSuccessMsg:
		m.currentView = viewShell
		m.output = fmt.Sprintf("Leave request created! (%s: %s to %s)", msg.leaveType, msg.startDate, msg.endDate)
		m.outputType = "success"
		return m, m.commandInput.Focus()
	}

	newModel, cmd := m.leaveRequestModel.Update(msg)
	m.leaveRequestModel = newModel.(LeaveRequestModel)

	if m.leaveRequestModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateLeaveList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.leaveListModel.Update(msg)
	m.leaveListModel = newModel.(LeaveListModel)

	if m.leaveListModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateLeaveCancel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.leaveCancelModel.Update(msg)
	m.leaveCancelModel = newModel.(LeaveCancelModel)

	if m.leaveCancelModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updatePeopleList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.peopleListModel.Update(msg)
	m.peopleListModel = newModel.(PeopleListModel)

	if m.peopleListModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateTeamList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" {
			m.currentView = viewShell
			return m, m.commandInput.Focus()
		}
	}

	newModel, cmd := m.teamListModel.Update(msg)
	m.teamListModel = newModel.(TeamListModel)

	if m.teamListModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) updateChangePassword(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.changePasswordModel.Update(msg)
	m.changePasswordModel = newModel.(ChangePasswordModel)

	if m.changePasswordModel.shouldGoBack {
		m.currentView = viewShell
		return m, m.commandInput.Focus()
	}

	return m, cmd
}

func (m ShellModel) getUserRole() string {
	if u := m.authService.GetUser(); u != nil {
		return u.Role
	}
	return ""
}

// View renders the shell
func (m ShellModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Render sub-model if active
	switch m.currentView {
	case viewDeviceFlow:
		return m.deviceFlowModel.View()
	case viewLogin:
		return m.loginModel.View()
	case viewRegister:
		return m.registerModel.View()
	case viewStandupSubmit:
		return m.standupSubmitModel.View()
	case viewStandupToday:
		return m.standupTodayModel.View()
	case viewStandupHistory:
		return m.standupHistoryModel.View()
	case viewAttendanceCheckIn:
		return m.attendanceCheckInModel.View()
	case viewAttendanceCheckOut:
		return m.attendanceCheckOutModel.View()
	case viewAttendanceToday:
		return m.attendanceTodayModel.View()
	case viewAttendanceHistory:
		return m.attendanceHistoryModel.View()
	case viewJoin:
		return m.joinModel.View()
	case viewLeaveRequest:
		return m.leaveRequestModel.View()
	case viewLeaveList:
		return m.leaveListModel.View()
	case viewLeaveCancel:
		return m.leaveCancelModel.View()
	case viewPeopleList:
		return m.peopleListModel.View()
	case viewTeamList:
		return m.teamListModel.View()
	case viewChangePassword:
		return m.changePasswordModel.View()
	}

	// Render shell view
	var b strings.Builder

	// Dashboard or output area
	if m.output == "" {
		if m.authService.IsAuthenticated() {
			b.WriteString(m.renderDashboard())
		} else {
			b.WriteString(m.renderUnauthDashboard())
		}
		b.WriteString("\n")
	}

	// Header + Output area (when output is shown, dashboard is replaced)
	if m.output != "" {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(mutedColor).
			Padding(0, 1)

		headerText := "Faro Helm"
		if m.authService.IsAuthenticated() {
			user := m.authService.GetUser()
			org := m.authService.GetOrganization()
			if user != nil && org != nil {
				headerText = fmt.Sprintf("Faro Helm  •  %s @ %s", user.Name, org.Name)
			}
		}
		b.WriteString(headerStyle.Render(headerText))
		b.WriteString("\n\n")
		var outputStyle lipgloss.Style
		switch m.outputType {
		case "success":
			outputStyle = successStyle
		case "error":
			outputStyle = errorStyle
		default:
			outputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		}
		b.WriteString(outputStyle.Render(m.output))
		b.WriteString("\n\n")
	}

	// Command input
	b.WriteString(m.commandInput.View())
	b.WriteString("\n\n")

	// Help footer
	footerStyle := lipgloss.NewStyle().Foreground(mutedColor)
	b.WriteString(footerStyle.Render("[Tab] Accept  [↑↓] Navigate  [Enter] Execute  [Ctrl+C] Quit"))

	return baseStyle.Render(b.String())
}

// Helper methods

func (m ShellModel) renderHelp() string {
	var b strings.Builder

	b.WriteString("Available Commands:\n\n")

	categories := m.registry.GetByCategory(m.authService.IsAuthenticated(), m.getUserRole())
	categoryOrder := []string{"auth", "standup", "attendance", "leave", "team", "people", "account", "utility"}
	// When authenticated, auth commands (logout) go near the end
	if m.authService.IsAuthenticated() {
		categoryOrder = []string{"standup", "attendance", "leave", "team", "people", "account", "auth", "utility"}
	}
	categoryNames := map[string]string{
		"auth":       "Authentication",
		"standup":    "Standups",
		"attendance": "Attendance",
		"leave":      "Leaves",
		"team":       "Teams",
		"people":     "People",
		"account":    "Account",
		"utility":    "Utility",
	}

	for _, cat := range categoryOrder {
		cmds, ok := categories[cat]
		if !ok || len(cmds) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("  %s\n", categoryNames[cat]))
		for _, cmd := range cmds {
			b.WriteString(fmt.Sprintf("    /%s - %s\n", cmd.Name, cmd.Description))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m ShellModel) renderWhoami() string {
	if !m.authService.IsAuthenticated() {
		return "Not logged in"
	}

	user := m.authService.GetUser()
	org := m.authService.GetOrganization()

	if user == nil || org == nil {
		return "User info not available"
	}

	return fmt.Sprintf("User: %s (%s)\nEmail: %s\nRole: %s\nOrganization: %s",
		user.Name, user.Role, user.Email, user.Role, org.Name)
}

// Dashboard rendering

func (m ShellModel) getGreeting() string {
	hour := time.Now().Hour()
	switch {
	case hour < 12:
		return "Good morning"
	case hour < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

// stripAnsi removes ANSI escape sequences to get the visible length of a string
func stripAnsi(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until we hit a letter
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++ // skip the final letter
			}
			i = j
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

// visibleLen returns the visible character count (excluding ANSI codes), counting runes
func visibleLen(s string) int {
	cleaned := stripAnsi(s)
	count := 0
	for range cleaned {
		count++
	}
	return count
}

// padToWidth pads a styled string to a target visible width with spaces
func padToWidth(s string, width int) string {
	vl := visibleLen(s)
	if vl >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vl)
}

// buildBoxedDashboard constructs a bordered box with optional two columns
// title goes in the top border, leftLines and rightLines are the column content
func buildBoxedDashboard(title string, leftLines, rightLines []string, totalWidth int, borderColor lipgloss.Color) string {
	border := lipgloss.NewStyle().Foreground(borderColor)

	// Calculate inner width (excluding the 2 border chars │ on each side)
	innerWidth := totalWidth - 2
	if innerWidth < 20 {
		innerWidth = 20
	}

	hasTwoCols := len(rightLines) > 0

	// Column widths
	var leftW, rightW int
	if hasTwoCols {
		leftW = innerWidth*2/5 - 1 // -1 for the separator │
		rightW = innerWidth - leftW - 1
	} else {
		leftW = innerWidth
	}

	// Top border: ╭─── Title ───...╮
	titleVis := visibleLen(title)
	topFill := innerWidth - 4 - titleVis - 1 // "─── " + title + " " + fill + "╮"
	if topFill < 1 {
		topFill = 1
	}
	topLine := border.Render("╭─── ") + title + border.Render(" "+strings.Repeat("─", topFill)+"╮")

	// Bottom border: ╰───...╯
	bottomLine := border.Render("╰" + strings.Repeat("─", innerWidth) + "╯")

	// Build rows
	maxRows := len(leftLines)
	if len(rightLines) > maxRows {
		maxRows = len(rightLines)
	}

	var rows []string
	rows = append(rows, topLine)

	for i := 0; i < maxRows; i++ {
		left := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		leftPadded := padToWidth(left, leftW)

		if hasTwoCols {
			right := ""
			if i < len(rightLines) {
				right = rightLines[i]
			}
			rightPadded := padToWidth(right, rightW)
			rows = append(rows, border.Render("│")+leftPadded+border.Render("│")+rightPadded+border.Render("│"))
		} else {
			rows = append(rows, border.Render("│")+leftPadded+border.Render("│"))
		}
	}

	rows = append(rows, bottomLine)
	return strings.Join(rows, "\n")
}

func (m ShellModel) renderDashboard() string {
	if m.dashboard.loading {
		spinnerStyle := lipgloss.NewStyle().Foreground(mutedColor)
		return spinnerStyle.Render("  Loading dashboard...")
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	title := titleStyle.Render("Faro Helm")

	leftLines := m.getDashboardLeftLines()
	rightLines := m.getDashboardRightLines()

	boxWidth := m.width - 4 // account for baseStyle padding
	if boxWidth < 40 {
		boxWidth = 40
	}

	if m.width < 80 {
		// Narrow: stack vertically, no right column
		allLines := append(leftLines, "")
		allLines = append(allLines, rightLines...)
		return buildBoxedDashboard(title, allLines, nil, boxWidth, mutedColor)
	}

	return buildBoxedDashboard(title, leftLines, rightLines, boxWidth, mutedColor)
}

func (m ShellModel) getDashboardLeftLines() []string {
	greetStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	dim := lipgloss.NewStyle().Foreground(mutedColor)

	userName := "there"
	role := ""
	orgName := ""
	if m.config.User != nil {
		userName = m.config.User.Name
		role = m.config.User.Role
	}
	if m.config.Organization != nil {
		orgName = m.config.Organization.Name
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, " "+greetStyle.Render(fmt.Sprintf("%s, %s!", m.getGreeting(), userName)))
	lines = append(lines, "")

	if role != "" && orgName != "" {
		lines = append(lines, "   "+dim.Render(role+" @ "+orgName))
	}
	lines = append(lines, "   "+dim.Render(time.Now().Format("Mon, Jan 2, 2006")))
	lines = append(lines, "")

	return lines
}

func (m ShellModel) getDashboardRightLines() []string {
	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(secondaryColor)
	dim := lipgloss.NewStyle().Foreground(mutedColor)
	normal := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	good := lipgloss.NewStyle().Foreground(successColor)
	warn := lipgloss.NewStyle().Foreground(secondaryColor)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, " "+sectionHeader.Render("Today's Snapshot"))

	if m.dashboard.myStandup != nil {
		lines = append(lines, fmt.Sprintf("   %s  %s", good.Render("✓"), normal.Render("Standup submitted")))
	} else {
		lines = append(lines, fmt.Sprintf("   %s  %s", warn.Render("○"), dim.Render("Standup not submitted")))
	}

	if m.dashboard.myAttendance != nil {
		lines = append(lines, fmt.Sprintf("   %s  %s", good.Render("✓"), normal.Render("Marked "+m.dashboard.myAttendance.Status)))
	} else {
		lines = append(lines, fmt.Sprintf("   %s  %s", warn.Render("○"), dim.Render("Attendance not marked")))
	}

	if len(m.dashboard.onLeaveToday) > 0 {
		names := strings.Join(m.dashboard.onLeaveToday, ", ")
		lines = append(lines, fmt.Sprintf("   %s  %s", dim.Render("✈"), dim.Render("On leave: "+names)))
	}

	lines = append(lines, "")

	return lines
}

func (m ShellModel) renderUnauthDashboard() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	title := titleStyle.Render("Faro Helm")

	cmdStyle := lipgloss.NewStyle().Foreground(primaryColor)
	hintStyle := lipgloss.NewStyle().Foreground(mutedColor)
	welcomeStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	hintHeader := lipgloss.NewStyle().Bold(true).Foreground(secondaryColor)

	var leftLines []string
	leftLines = append(leftLines, "")
	leftLines = append(leftLines, " "+welcomeStyle.Render("Welcome to Faro Helm!"))
	leftLines = append(leftLines, "")
	leftLines = append(leftLines, " "+descStyle.Render("Standups, attendance,"))
	leftLines = append(leftLines, " "+descStyle.Render("and leave management."))
	leftLines = append(leftLines, "")

	var rightLines []string
	rightLines = append(rightLines, "")
	rightLines = append(rightLines, " "+hintHeader.Render("Get Started"))

	hints := []struct{ cmd, desc string }{
		{"/login", "Sign in to your account"},
		{"/signup", "Create a new organization"},
		{"/join", "Join an existing organization"},
		{"/help", "View all available commands"},
	}
	for _, h := range hints {
		rightLines = append(rightLines, fmt.Sprintf("   %s  %s", cmdStyle.Render(h.cmd), hintStyle.Render(h.desc)))
	}
	rightLines = append(rightLines, "")

	boxWidth := m.width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	if m.width < 80 {
		allLines := append(leftLines, rightLines...)
		return buildBoxedDashboard(title, allLines, nil, boxWidth, mutedColor)
	}

	return buildBoxedDashboard(title, leftLines, rightLines, boxWidth, mutedColor)
}


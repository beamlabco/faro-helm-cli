package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/auth"
)

// DeviceFlowModel handles the OAuth device flow login screen.
type DeviceFlowModel struct {
	authService *auth.Service
	onSuccess   func()

	// Flow state
	deviceCode      string
	userCode        string
	verificationURI string
	interval        int // seconds between polls

	phase    string // "loading" | "waiting" | "success" | "error"
	errorMsg string
	quitting bool

	// Spinner frame
	spinnerFrame int
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewDeviceFlowModel creates a new device flow model.
func NewDeviceFlowModel(authService *auth.Service, onSuccess func()) DeviceFlowModel {
	return DeviceFlowModel{
		authService: authService,
		onSuccess:   onSuccess,
		phase:       "loading",
		interval:    5,
	}
}

// Init kicks off device flow initiation.
func (m DeviceFlowModel) Init() tea.Cmd {
	return m.initiateDeviceFlow()
}

// Message types

type deviceFlowInitMsg struct {
	deviceCode      string
	userCode        string
	verificationURI string
	interval        int
}

type deviceFlowInitErrMsg string

type deviceFlowPollMsg struct {
	tokenPair *api.DeviceTokenPair
	err       error
}

type deviceFlowTickMsg struct{}

// Update handles messages.
func (m DeviceFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.quitting = true
			return m, nil
		}

	case deviceFlowInitMsg:
		m.deviceCode = msg.deviceCode
		m.userCode = msg.userCode
		m.verificationURI = msg.verificationURI
		m.interval = msg.interval
		m.phase = "waiting"
		return m, m.schedulePoll()

	case deviceFlowInitErrMsg:
		m.phase = "error"
		m.errorMsg = string(msg)
		return m, nil

	case deviceFlowTickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		return m, m.poll()

	case deviceFlowPollMsg:
		if msg.err != nil {
			switch msg.err.Error() {
			case "authorization_pending":
				return m, m.schedulePoll()
			case "slow_down":
				m.interval += 5
				return m, m.schedulePoll()
			case "expired_token":
				m.phase = "error"
				m.errorMsg = "Device code expired. Press Esc and try /login again."
				return m, nil
			case "access_denied":
				m.phase = "error"
				m.errorMsg = "Authorization denied in browser."
				return m, nil
			default:
				m.phase = "error"
				m.errorMsg = fmt.Sprintf("Authentication failed: %s", msg.err.Error())
				return m, nil
			}
		}

		// Success — save auth data
		if err := m.authService.CompleteDeviceLogin(msg.tokenPair); err != nil {
			m.phase = "error"
			m.errorMsg = fmt.Sprintf("Login failed: %s", err.Error())
			return m, nil
		}

		m.phase = "success"
		if m.onSuccess != nil {
			m.onSuccess()
		}
		return m, deviceFlowSuccessCmd()
	}

	return m, nil
}

// View renders the device flow screen.
func (m DeviceFlowModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Sign in to Faro Helm"))
	b.WriteString("\n\n")

	switch m.phase {
	case "loading":
		spinner := spinnerFrames[m.spinnerFrame]
		b.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Render(spinner + "  Starting device flow…"))

	case "waiting":
		codeStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

		b.WriteString(labelStyle.Render("1. Open this URL in your browser:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Bold(true).Render("   " + m.verificationURI))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("2. Enter this code:"))
		b.WriteString("\n")
		b.WriteString(codeStyle.Render(m.userCode))
		b.WriteString("\n")

		spinner := spinnerFrames[m.spinnerFrame]
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(
			fmt.Sprintf("%s  Waiting for approval…", spinner),
		))

	case "success":
		b.WriteString(successStyle.Render("✓ Signed in successfully!"))

	case "error":
		b.WriteString(errorStyle.Render("✗ " + m.errorMsg))
	}

	b.WriteString("\n\n")
	if m.phase == "waiting" || m.phase == "loading" {
		b.WriteString(helpStyle.Render("esc: cancel"))
	}

	return baseStyle.Render(b.String())
}

// initiateDeviceFlow starts the device authorization flow.
func (m *DeviceFlowModel) initiateDeviceFlow() tea.Cmd {
	return func() tea.Msg {
		info, err := m.authService.StartDeviceFlow()
		if err != nil {
			return deviceFlowInitErrMsg(err.Error())
		}
		return deviceFlowInitMsg{
			deviceCode:      info.DeviceCode,
			userCode:        info.UserCode,
			verificationURI: info.VerificationURI,
			interval:        info.Interval,
		}
	}
}

// schedulePoll waits the poll interval then triggers a spinner tick + poll.
func (m *DeviceFlowModel) schedulePoll() tea.Cmd {
	interval := m.interval
	if interval <= 0 {
		interval = 5
	}
	return tea.Tick(time.Duration(interval)*time.Second, func(_ time.Time) tea.Msg {
		return deviceFlowTickMsg{}
	})
}

// poll calls the token endpoint once.
func (m *DeviceFlowModel) poll() tea.Cmd {
	deviceCode := m.deviceCode
	return func() tea.Msg {
		tokenPair, err := m.authService.PollDeviceToken(deviceCode)
		return deviceFlowPollMsg{tokenPair: tokenPair, err: err}
	}
}

func deviceFlowSuccessCmd() tea.Cmd {
	return func() tea.Msg {
		return deviceFlowSuccessMsg{}
	}
}

type deviceFlowSuccessMsg struct{}

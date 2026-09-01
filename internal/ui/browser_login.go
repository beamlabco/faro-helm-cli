package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/beamlabco/faro-helm-cli/internal/auth"
	"github.com/beamlabco/faro-helm-cli/internal/oauthflow"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// BrowserLoginModel handles the Authorization Code + PKCE login screen:
// opens the system browser to /oauth/authorize and waits for the local
// loopback callback. Replaces the old device-flow (code-entry) screen —
// the CLI always has a browser available, so there's no need for a
// separate browserless path.
type BrowserLoginModel struct {
	authService *auth.Service
	onSuccess   func()

	bl     *auth.BrowserLogin
	cancel context.CancelFunc

	phase        string // "starting" | "waiting" | "success" | "error"
	errorMsg     string
	quitting     bool
	spinnerFrame int
}

// NewBrowserLoginModel creates a new browser login model.
func NewBrowserLoginModel(authService *auth.Service, onSuccess func()) BrowserLoginModel {
	return BrowserLoginModel{
		authService: authService,
		onSuccess:   onSuccess,
		phase:       "starting",
	}
}

// Init kicks off the flow.
func (m BrowserLoginModel) Init() tea.Cmd {
	return tea.Batch(m.begin(), tickBrowserLoginSpinner())
}

// Message types

type browserLoginBeganMsg struct{ bl *auth.BrowserLogin }
type browserLoginErrMsg string
type browserLoginDoneMsg struct{ err error }
type browserLoginTickMsg struct{}
type browserLoginSuccessMsg struct{}

func (m *BrowserLoginModel) begin() tea.Cmd {
	return func() tea.Msg {
		bl, err := m.authService.BeginBrowserLogin()
		if err != nil {
			return browserLoginErrMsg(err.Error())
		}
		return browserLoginBeganMsg{bl: bl}
	}
}

// awaitCompletion blocks (inside the returned Cmd's goroutine) until the
// browser redirect arrives, is cancelled via m.cancel, or times out.
func (m *BrowserLoginModel) awaitCompletion() tea.Cmd {
	bl := m.bl
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	return func() tea.Msg {
		err := m.authService.CompleteBrowserLogin(ctx, bl)
		return browserLoginDoneMsg{err: err}
	}
}

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		_ = oauthflow.OpenBrowser(url) // best-effort — the URL is also shown on screen
		return nil
	}
}

func tickBrowserLoginSpinner() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(_ time.Time) tea.Msg {
		return browserLoginTickMsg{}
	})
}

// Update handles messages.
func (m BrowserLoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.cancel != nil {
				m.cancel()
			}
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.cancel != nil {
				m.cancel()
			}
			m.quitting = true
			return m, nil
		}

	case browserLoginBeganMsg:
		m.bl = msg.bl
		m.phase = "waiting"
		return m, tea.Batch(openBrowserCmd(msg.bl.AuthorizeURL), m.awaitCompletion())

	case browserLoginErrMsg:
		m.phase = "error"
		m.errorMsg = string(msg)
		return m, nil

	case browserLoginTickMsg:
		if m.phase == "success" || m.phase == "error" {
			return m, nil
		}
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		return m, tickBrowserLoginSpinner()

	case browserLoginDoneMsg:
		if msg.err != nil {
			m.phase = "error"
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.phase = "success"
		if m.onSuccess != nil {
			m.onSuccess()
		}
		return m, func() tea.Msg { return browserLoginSuccessMsg{} }
	}

	return m, nil
}

// View renders the browser login screen.
func (m BrowserLoginModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Sign in to Faro Helm"))
	b.WriteString("\n\n")

	switch m.phase {
	case "starting":
		spinner := spinnerFrames[m.spinnerFrame]
		b.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Render(spinner + "  Starting sign-in…"))

	case "waiting":
		b.WriteString(labelStyle.Render("Opening your browser to sign in…"))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("If it doesn't open automatically, visit:"))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Bold(true).Render("   " + m.bl.AuthorizeURL))
		b.WriteString("\n\n")

		spinner := spinnerFrames[m.spinnerFrame]
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(
			fmt.Sprintf("%s  Waiting for you to finish signing in…", spinner),
		))

	case "success":
		b.WriteString(successStyle.Render("✓ Signed in successfully!"))

	case "error":
		b.WriteString(errorStyle.Render("✗ " + m.errorMsg))
	}

	b.WriteString("\n\n")
	if m.phase == "waiting" || m.phase == "starting" {
		b.WriteString(helpStyle.Render("esc: cancel"))
	}

	return baseStyle.Render(b.String())
}

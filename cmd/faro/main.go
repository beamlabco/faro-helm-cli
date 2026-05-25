package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/beamlabco/faro-helm/internal/api"
	"github.com/beamlabco/faro-helm/internal/attendance"
	"github.com/beamlabco/faro-helm/internal/auth"
	"github.com/beamlabco/faro-helm/internal/config"
	"github.com/beamlabco/faro-helm/internal/invitation"
	"github.com/beamlabco/faro-helm/internal/leave"
	"github.com/beamlabco/faro-helm/internal/organization"
	"github.com/beamlabco/faro-helm/internal/project"
	"github.com/beamlabco/faro-helm/internal/standup"
	"github.com/beamlabco/faro-helm/internal/user"
	"github.com/beamlabco/faro-helm/internal/ui"
)

var (
	version             = "dev"
	defaultBaseURL      = "http://localhost:3001"
	defaultAuthBaseURL  = "http://localhost:3000"
)

func main() {
	if defaultBaseURL != "" {
		config.DefaultBaseURL = defaultBaseURL
	}
	if defaultAuthBaseURL != "" {
		config.DefaultAuthBaseURL = defaultAuthBaseURL
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	apiClient := api.NewClientFromConfig(cfg)

	authAPIURL := cfg.API.AuthBaseURL
	authAPIClient := api.NewAuthClient(authAPIURL)

	authService := auth.NewService(apiClient, authAPIClient, cfg)
	standupService := standup.NewService(apiClient)
	attendanceService := attendance.NewService(apiClient)
	invitationService := invitation.NewService(apiClient)
	leaveService := leave.NewService(apiClient)
	userService := user.NewService(apiClient)
	orgService := organization.NewService(apiClient)
	projectService := project.NewService(apiClient)

	ui.PrintLogo()

	shellModel := ui.NewShellModel(authService, standupService, attendanceService, invitationService, leaveService, userService, orgService, projectService, cfg)
	p := tea.NewProgram(shellModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

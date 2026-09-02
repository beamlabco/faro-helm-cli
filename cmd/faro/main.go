package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/beamlabco/faro-helm-cli/internal/api"
	"github.com/beamlabco/faro-helm-cli/internal/attendance"
	"github.com/beamlabco/faro-helm-cli/internal/auth"
	"github.com/beamlabco/faro-helm-cli/internal/config"
	"github.com/beamlabco/faro-helm-cli/internal/leave"
	"github.com/beamlabco/faro-helm-cli/internal/project"
	"github.com/beamlabco/faro-helm-cli/internal/standup"
	"github.com/beamlabco/faro-helm-cli/internal/user"
	"github.com/beamlabco/faro-helm-cli/internal/ui"
)

var (
	version        = "dev"
	defaultBaseURL = "http://localhost:3001"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(version)
		return
	}

	if defaultBaseURL != "" {
		config.DefaultBaseURL = defaultBaseURL
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	apiClient := api.NewClientFromConfig(cfg)

	authService := auth.NewService(apiClient, cfg)
	standupService := standup.NewService(apiClient)
	attendanceService := attendance.NewService(apiClient)
	leaveService := leave.NewService(apiClient)
	userService := user.NewService(apiClient, func() string {
		if cfg.Auth == nil {
			return ""
		}
		return cfg.Auth.Token
	})
	projectService := project.NewService(apiClient)

	shellModel := ui.NewShellModel(authService, standupService, attendanceService, leaveService, userService, projectService, cfg)
	p := tea.NewProgram(shellModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#1D9E75")
	secondaryColor = lipgloss.Color("#5DCAA5")
	errorColor     = lipgloss.Color("#EF4444")
	mutedColor     = lipgloss.Color("#6B7280")
	successColor   = lipgloss.Color("#10B981")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Input label style
	labelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Focused input style
	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Blurred input style
	blurredInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(mutedColor).
				Padding(0, 1)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			MarginTop(1)

	// Success message style
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			MarginTop(1)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(2)

	// Button style
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 3).
			MarginTop(1)

	// Secondary button style
	secondaryButtonStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 3).
				MarginTop(1)
)

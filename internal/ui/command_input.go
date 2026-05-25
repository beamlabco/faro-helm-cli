package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandInput is a text input with autocomplete support
type CommandInput struct {
	input       textinput.Model
	suggestions *Suggestions
	registry    *CommandRegistry
	isAuth      bool
	role        string
	ghostText   string
}

// NewCommandInput creates a new command input
func NewCommandInput(registry *CommandRegistry, isAuth bool, role string) *CommandInput {
	ti := textinput.New()
	ti.Placeholder = "Type / for commands..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	return &CommandInput{
		input:       ti,
		suggestions: NewSuggestions(),
		registry:    registry,
		isAuth:      isAuth,
		role:        role,
		ghostText:   "",
	}
}

// SetAuth updates the authentication state and role
func (c *CommandInput) SetAuth(isAuth bool, role string) {
	c.isAuth = isAuth
	c.role = role
}

// Focus focuses the input
func (c *CommandInput) Focus() tea.Cmd {
	return c.input.Focus()
}

// Blur blurs the input
func (c *CommandInput) Blur() {
	c.input.Blur()
}

// Value returns the current input value
func (c *CommandInput) Value() string {
	return c.input.Value()
}

// SetValue sets the input value
func (c *CommandInput) SetValue(v string) {
	c.input.SetValue(v)
	c.updateSuggestions()
}

// Reset clears the input
func (c *CommandInput) Reset() {
	c.input.SetValue("")
	c.suggestions.Clear()
	c.ghostText = ""
}

// SuggestionsVisible returns whether suggestions are showing
func (c *CommandInput) SuggestionsVisible() bool {
	return c.suggestions.IsVisible()
}

// updateSuggestions updates the suggestion list based on input
func (c *CommandInput) updateSuggestions() {
	value := c.input.Value()

	if strings.HasPrefix(value, "/") {
		matches := c.registry.Search(value, c.isAuth, c.role)
		c.suggestions.SetItems(matches)

		// Update ghost text
		if first := c.suggestions.First(); first != nil {
			query := strings.TrimPrefix(value, "/")
			if strings.HasPrefix(first.Name, query) && len(first.Name) > len(query) {
				c.ghostText = first.Name[len(query):]
			} else {
				c.ghostText = ""
			}
		} else {
			c.ghostText = ""
		}
	} else {
		c.suggestions.Clear()
		c.ghostText = ""
	}
}

// AcceptSuggestion accepts the current suggestion
func (c *CommandInput) AcceptSuggestion() bool {
	if selected := c.suggestions.Selected(); selected != nil {
		c.input.SetValue("/" + selected.Name)
		c.suggestions.Clear()
		c.ghostText = ""
		return true
	}
	return false
}

// AcceptGhostText accepts the ghost text completion
func (c *CommandInput) AcceptGhostText() bool {
	if c.ghostText != "" {
		c.input.SetValue(c.input.Value() + c.ghostText)
		c.ghostText = ""
		c.updateSuggestions()
		return true
	}
	return false
}

// SelectedCommand returns the currently selected command from dropdown (if any)
func (c *CommandInput) SelectedCommand() *Command {
	if c.suggestions.IsVisible() {
		return c.suggestions.Selected()
	}
	return nil
}

// Update handles input updates
func (c *CommandInput) Update(msg tea.Msg) (*CommandInput, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if c.suggestions.IsVisible() {
				c.suggestions.MoveUp()
				return c, nil
			}
		case "down":
			if c.suggestions.IsVisible() {
				c.suggestions.MoveDown()
				return c, nil
			}
		case "tab":
			// Accept ghost text or selected suggestion
			if c.ghostText != "" {
				c.AcceptGhostText()
				return c, nil
			} else if c.suggestions.IsVisible() {
				c.AcceptSuggestion()
				return c, nil
			}
		case "esc":
			if c.suggestions.IsVisible() {
				c.suggestions.Hide()
				return c, nil
			}
		}
	}

	// Update the text input
	c.input, cmd = c.input.Update(msg)
	c.updateSuggestions()

	return c, cmd
}

// View renders the command input with ghost text and suggestions
func (c *CommandInput) View() string {
	var b strings.Builder

	// Render input with ghost text
	inputView := c.input.View()

	if c.ghostText != "" {
		ghostStyle := lipgloss.NewStyle().Foreground(mutedColor)
		// Insert ghost text after cursor position
		inputView += ghostStyle.Render(c.ghostText)
	}

	b.WriteString(inputView)

	// Render suggestions below input
	if c.suggestions.IsVisible() {
		b.WriteString("\n")
		b.WriteString(c.suggestions.View())
	}

	return b.String()
}

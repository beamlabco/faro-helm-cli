package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxVisibleSuggestions = 6

// Suggestions represents the dropdown suggestion component
type Suggestions struct {
	items    []Command
	selected int
	visible  bool
}

// NewSuggestions creates a new suggestions component
func NewSuggestions() *Suggestions {
	return &Suggestions{
		items:    nil,
		selected: 0,
		visible:  false,
	}
}

// SetItems updates the suggestion items
func (s *Suggestions) SetItems(items []Command) {
	// Only reset selection if items actually changed
	if !commandsEqual(s.items, items) {
		s.items = items
		s.selected = 0
	}
	s.visible = len(items) > 0
}

// commandsEqual checks if two command slices are equal
func commandsEqual(a, b []Command) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			return false
		}
	}
	return true
}

// Clear hides and clears suggestions
func (s *Suggestions) Clear() {
	s.items = nil
	s.selected = 0
	s.visible = false
}

// IsVisible returns whether suggestions are visible
func (s *Suggestions) IsVisible() bool {
	return s.visible && len(s.items) > 0
}

// Show makes suggestions visible
func (s *Suggestions) Show() {
	s.visible = len(s.items) > 0
}

// Hide hides suggestions
func (s *Suggestions) Hide() {
	s.visible = false
}

// MoveUp moves selection up
func (s *Suggestions) MoveUp() {
	if len(s.items) == 0 {
		return
	}
	s.selected--
	if s.selected < 0 {
		s.selected = len(s.items) - 1
	}
}

// MoveDown moves selection down
func (s *Suggestions) MoveDown() {
	if len(s.items) == 0 {
		return
	}
	s.selected++
	if s.selected >= len(s.items) {
		s.selected = 0
	}
}

// Selected returns the currently selected command
func (s *Suggestions) Selected() *Command {
	if len(s.items) == 0 || s.selected >= len(s.items) {
		return nil
	}
	return &s.items[s.selected]
}

// First returns the first suggestion (for ghost text)
func (s *Suggestions) First() *Command {
	if len(s.items) == 0 {
		return nil
	}
	return &s.items[0]
}

// View renders the suggestions dropdown
func (s *Suggestions) View() string {
	if !s.visible || len(s.items) == 0 {
		return ""
	}

	var b strings.Builder

	// Determine visible range
	visibleCount := len(s.items)
	if visibleCount > maxVisibleSuggestions {
		visibleCount = maxVisibleSuggestions
	}

	// Calculate scroll offset to keep selected item visible
	startIdx := 0
	if s.selected >= maxVisibleSuggestions {
		startIdx = s.selected - maxVisibleSuggestions + 1
	}
	endIdx := startIdx + visibleCount
	if endIdx > len(s.items) {
		endIdx = len(s.items)
		startIdx = endIdx - visibleCount
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Styles
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	descStyle := lipgloss.NewStyle().
		Foreground(mutedColor)

	// Find max command width for alignment
	maxCmdWidth := 0
	for _, cmd := range s.items[startIdx:endIdx] {
		cmdStr := "/" + cmd.Name
		if len(cmdStr) > maxCmdWidth {
			maxCmdWidth = len(cmdStr)
		}
	}

	// Render items
	for i := startIdx; i < endIdx; i++ {
		cmd := s.items[i]
		cmdStr := "/" + cmd.Name
		padding := strings.Repeat(" ", maxCmdWidth-len(cmdStr)+2)

		line := fmt.Sprintf("%s%s%s", cmdStr, padding, cmd.Description)

		if i == s.selected {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(cmdStr + padding))
			b.WriteString(descStyle.Render(cmd.Description))
		}

		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Show scroll indicator if there are more items
	if len(s.items) > maxVisibleSuggestions {
		scrollInfo := fmt.Sprintf(" (%d/%d)", s.selected+1, len(s.items))
		b.WriteString("\n")
		b.WriteString(descStyle.Render(scrollInfo))
	}

	return boxStyle.Render(b.String())
}

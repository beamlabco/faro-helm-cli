package ui

import (
	"sort"
	"strings"
)

// Command represents a slash command
type Command struct {
	Name         string
	Aliases      []string
	Description  string
	Category     string
	RequiresAuth bool
	HideWhenAuth bool
	RequiresRole string // if set, only users with this role see the command
}

// CommandRegistry holds all available commands
type CommandRegistry struct {
	commands []Command
}

// NewCommandRegistry creates a new registry with all commands
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: []Command{
			// Auth (unauthenticated)
			{Name: "login", Aliases: []string{"signin"}, Description: "Sign in to your account", Category: "auth", RequiresAuth: false, HideWhenAuth: true},
			{Name: "signup", Aliases: []string{"register"}, Description: "Create a new account", Category: "auth", RequiresAuth: false, HideWhenAuth: true},
			{Name: "join", Aliases: nil, Description: "Join an organization via invitation", Category: "auth", RequiresAuth: false, HideWhenAuth: true},

			// Standups
			{Name: "standup", Aliases: nil, Description: "Submit your daily standup", Category: "standup", RequiresAuth: true},
			{Name: "standup today", Aliases: nil, Description: "View your standup for today", Category: "standup", RequiresAuth: true},
			{Name: "standup history", Aliases: []string{"standup my"}, Description: "View your standup history", Category: "standup", RequiresAuth: true},

			// Attendance
			{Name: "checkin", Aliases: []string{"in"}, Description: "Check in for the day", Category: "attendance", RequiresAuth: true},
			{Name: "checkout", Aliases: []string{"out"}, Description: "Check out for the day", Category: "attendance", RequiresAuth: true},
			{Name: "attendance today", Aliases: nil, Description: "View your attendance for today", Category: "attendance", RequiresAuth: true},
			{Name: "attendance history", Aliases: []string{"attendance my"}, Description: "View your attendance history", Category: "attendance", RequiresAuth: true},

			// Leaves
			{Name: "leave", Aliases: nil, Description: "Request a leave", Category: "leave", RequiresAuth: true},
			{Name: "leave list", Aliases: []string{"leaves"}, Description: "View team leaves", Category: "leave", RequiresAuth: true},
			{Name: "leave cancel", Aliases: nil, Description: "Cancel a pending leave", Category: "leave", RequiresAuth: true},

			// Teams
			{Name: "team", Aliases: []string{"teams"}, Description: "View your teams", Category: "team", RequiresAuth: true},

			// People
			{Name: "people", Aliases: []string{"members"}, Description: "View workspace people and roles", Category: "people", RequiresAuth: true},

			// Account
			{Name: "password", Aliases: []string{"change-password"}, Description: "Change your password", Category: "account", RequiresAuth: true},

			// Utility
			{Name: "help", Aliases: []string{"?"}, Description: "Show available commands", Category: "utility", RequiresAuth: false},
			{Name: "whoami", Aliases: []string{"me"}, Description: "Show current user info", Category: "utility", RequiresAuth: true},
			{Name: "clear", Aliases: []string{"cls"}, Description: "Clear the screen", Category: "utility", RequiresAuth: false},
			{Name: "logout", Aliases: []string{"signout"}, Description: "Sign out of your account", Category: "auth", RequiresAuth: true},
			{Name: "quit", Aliases: []string{"exit", "q"}, Description: "Exit the application", Category: "utility", RequiresAuth: false},
		},
	}
}

// GetAll returns all commands
func (r *CommandRegistry) GetAll() []Command {
	return r.commands
}

// GetAvailable returns commands based on auth state and role
func (r *CommandRegistry) GetAvailable(isAuthenticated bool, role string) []Command {
	var available []Command
	for _, cmd := range r.commands {
		if cmd.RequiresAuth && !isAuthenticated {
			continue
		}
		if cmd.HideWhenAuth && isAuthenticated {
			continue
		}
		if cmd.RequiresRole != "" && cmd.RequiresRole != role {
			continue
		}
		available = append(available, cmd)
	}
	return available
}

// Search finds commands matching the query (prefix match)
func (r *CommandRegistry) Search(query string, isAuthenticated bool, role string) []Command {
	query = strings.ToLower(strings.TrimPrefix(query, "/"))
	if query == "" {
		return r.GetAvailable(isAuthenticated, role)
	}

	var matches []Command
	available := r.GetAvailable(isAuthenticated, role)

	for _, cmd := range available {
		// Check command name
		if strings.HasPrefix(strings.ToLower(cmd.Name), query) {
			matches = append(matches, cmd)
			continue
		}
		// Check aliases
		for _, alias := range cmd.Aliases {
			if strings.HasPrefix(strings.ToLower(alias), query) {
				matches = append(matches, cmd)
				break
			}
		}
	}

	// Sort by relevance (exact prefix matches first, then by name length)
	sort.Slice(matches, func(i, j int) bool {
		iExact := strings.HasPrefix(strings.ToLower(matches[i].Name), query)
		jExact := strings.HasPrefix(strings.ToLower(matches[j].Name), query)
		if iExact != jExact {
			return iExact
		}
		return len(matches[i].Name) < len(matches[j].Name)
	})

	return matches
}

// FindExact finds a command by exact name or alias
func (r *CommandRegistry) FindExact(name string) *Command {
	name = strings.ToLower(strings.TrimPrefix(name, "/"))

	for _, cmd := range r.commands {
		if strings.ToLower(cmd.Name) == name {
			return &cmd
		}
		for _, alias := range cmd.Aliases {
			if strings.ToLower(alias) == name {
				return &cmd
			}
		}
	}
	return nil
}

// GetByCategory returns commands grouped by category
func (r *CommandRegistry) GetByCategory(isAuthenticated bool, role string) map[string][]Command {
	result := make(map[string][]Command)
	for _, cmd := range r.GetAvailable(isAuthenticated, role) {
		result[cmd.Category] = append(result[cmd.Category], cmd)
	}
	return result
}

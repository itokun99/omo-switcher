package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all Lipgloss styles for the TUI.
type Styles struct {
	// Colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color

	// Component styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Active   lipgloss.Style
	Inactive lipgloss.Style
	Status   lipgloss.Style
	Help     lipgloss.Style
	ErrorMsg lipgloss.Style
	Border   lipgloss.Style
}

// DefaultStyles returns the default styles.
func DefaultStyles() Styles {
	return Styles{
		Primary:   lipgloss.Color("60"),  // Purple
		Secondary: lipgloss.Color("229"), // Light yellow
		Accent:    lipgloss.Color("212"), // Pink
		Error:     lipgloss.Color("196"), // Red
		Muted:     lipgloss.Color("241"), // Gray

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("60")),

		Subtitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")),

		Active: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42")), // Green

		Inactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		Status: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("60")).
			Foreground(lipgloss.Color("230")),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		ErrorMsg: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("60")),
	}
}

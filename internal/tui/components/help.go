// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpModel displays a help overlay with key bindings.
type HelpModel struct {
	active bool
	width  int
	height int
}

// NewHelpModel creates a HelpModel.
func NewHelpModel() HelpModel {
	return HelpModel{}
}

// Toggle switches the help overlay on/off.
func (m *HelpModel) Toggle() {
	m.active = !m.active
}

// Hide deactivates the help overlay.
func (m *HelpModel) Hide() {
	m.active = false
}

// IsActive reports whether help is active.
func (m HelpModel) IsActive() bool {
	return m.active
}

// SetSize updates the dimensions.
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Render returns the help overlay string.
func (m HelpModel) Render(styles HelpStyles) string {
	if !m.active {
		return ""
	}

	content := `  Key Bindings

  Navigation
    ↑/k        Move up
    ↓/j        Move down
    g/home     Jump to top
    G/end      Jump to bottom

  Actions
    Enter      Switch to config
    s          Show config detail
    /          Search configs
    ?          Toggle this help

  Utilities
    v          Validate all configs
    b          Backup manager
    d          Diff viewer
    i          Config info
    r          Reload configs

  General
    q/Esc      Quit`

	box := styles.Box.Render(content)

	boxWidth := lipgloss.Width(box)
	padding := max(0, (m.width-boxWidth)/2)

	topPadding := max(0, (m.height-lipgloss.Height(box))/2)

	var b strings.Builder
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(box)

	return b.String()
}

// HelpStyles holds styles for the help overlay.
type HelpStyles struct {
	Box lipgloss.Style
}

// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusModel displays the current config and key hints.
type StatusModel struct {
	activeAlias string
	activeGroup string
	message     string // transient message (success/error)
	loading     bool
	width       int
}

// NewStatusModel creates a StatusModel.
func NewStatusModel() StatusModel {
	return StatusModel{}
}

// SetActive updates the currently active config display.
func (m *StatusModel) SetActive(alias, group string) {
	m.activeAlias = alias
	m.activeGroup = group
}

// SetMessage sets a transient message (shown briefly then cleared).
func (m *StatusModel) SetMessage(msg string) {
	m.message = msg
}

// ClearMessage clears the transient message.
func (m *StatusModel) ClearMessage() {
	m.message = ""
}

// Message returns the current transient message.
func (m StatusModel) Message() string {
	return m.message
}

// SetWidth updates the width for rendering.
func (m *StatusModel) SetWidth(width int) {
	m.width = width
}

// Render returns the status bar string.
func (m StatusModel) Render(styles StatusStyles) string {
	var left, right string

	// Left side: active config or message
	if m.message != "" {
		left = m.message
	} else if m.activeAlias != "" {
		left = fmt.Sprintf("  %s (%s)", m.activeAlias, m.activeGroup)
	} else {
		left = "  No active config"
	}

	// Right side: key hints
	right = "j/k: navigate  enter: switch  /: search  ?: help  q: quit"

	// Calculate padding
	padding := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right))

	return styles.Bar.Render(
		left + strings.Repeat(" ", padding) + right,
	)
}

// StatusStyles holds styles for the status bar.
type StatusStyles struct {
	Bar lipgloss.Style
}

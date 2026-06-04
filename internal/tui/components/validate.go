// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ValidationResult holds the validation status of a config.
type ValidationResult struct {
	Alias   string
	IsValid bool
	Error   string
}

// ValidateModel displays validation results for all configs.
type ValidateModel struct {
	results []ValidationResult
	cursor  int
	active  bool
	height  int
}

// NewValidateModel creates a ValidateModel.
func NewValidateModel() ValidateModel {
	return ValidateModel{}
}

// Show activates the validator with results from the service.
func (m *ValidateModel) Show(results []ValidationResult, height int) {
	m.results = results
	m.cursor = 0
	m.height = height - 4
	m.active = true
}

// Hide deactivates the validator.
func (m *ValidateModel) Hide() {
	m.active = false
}

// IsActive reports whether validator is active.
func (m ValidateModel) IsActive() bool {
	return m.active
}

// MoveUp moves cursor up.
func (m *ValidateModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveDown moves cursor down.
func (m *ValidateModel) MoveDown() {
	if m.cursor < len(m.results)-1 {
		m.cursor++
	}
}

// Render returns the validator view string.
func (m ValidateModel) Render(styles ValidateStyles, width int) string {
	var b strings.Builder

	b.WriteString("  Config Validation\n\n")

	if len(m.results) == 0 {
		b.WriteString("  No configs to validate")
		return b.String()
	}

	for i, r := range m.results {
		cursor := "  "
		if i == m.cursor {
			cursor = "❯ "
		}

		status := "✓"
		statusStyle := styles.Valid
		if !r.IsValid {
			status = "✗"
			statusStyle = styles.Invalid
		}

		line := fmt.Sprintf("%s%s %s", cursor, status, r.Alias)
		b.WriteString(statusStyle.Render(line))

		if !r.IsValid && r.Error != "" {
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(styles.Error.Render(r.Error))
		}

		b.WriteString("\n")
	}

	return b.String()
}

// ValidateStyles holds styles for the validator.
type ValidateStyles struct {
	Valid   lipgloss.Style
	Invalid lipgloss.Style
	Error   lipgloss.Style
}

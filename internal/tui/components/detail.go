// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DetailModel displays config JSON content with scrolling support.
type DetailModel struct {
	alias    string
	filename string
	content  string
	lines    []string
	offset   int
	height   int
	width    int
	active   bool
}

// NewDetailModel creates a DetailModel.
func NewDetailModel() DetailModel {
	return DetailModel{}
}

// Show activates the detail view for a config.
func (m *DetailModel) Show(alias, filename, content string, height int) {
	m.alias = alias
	m.filename = filename
	m.content = content
	m.lines = strings.Split(content, "\n")
	m.offset = 0
	m.height = height - 4 // reserve space for header/footer
	m.active = true
}

// Hide deactivates the detail view.
func (m *DetailModel) Hide() {
	m.active = false
}

// IsActive reports whether detail view is active.
func (m DetailModel) IsActive() bool {
	return m.active
}

// ScrollUp scrolls the content up one line.
func (m *DetailModel) ScrollUp() {
	if m.offset > 0 {
		m.offset--
	}
}

// ScrollDown scrolls the content down one line.
func (m *DetailModel) ScrollDown() {
	maxOffset := max(len(m.lines)-m.height, 0)
	if m.offset < maxOffset {
		m.offset++
	}
}

// ScrollToTop scrolls to the beginning.
func (m *DetailModel) ScrollToTop() {
	m.offset = 0
}

// ScrollToBottom scrolls to the end.
func (m *DetailModel) ScrollToBottom() {
	m.offset = max(len(m.lines)-m.height, 0)
}

// SetWidth sets the display width.
func (m *DetailModel) SetWidth(width int) {
	m.width = width
}

// Render returns the detail view string.
func (m DetailModel) Render(styles DetailStyles, width int) string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %s (%s)", m.alias, m.filename)
	b.WriteString(styles.Header.Render(header))
	b.WriteString("\n\n")

	// Content with scrolling
	end := min(m.offset+m.height, len(m.lines))

	for i := m.offset; i < end; i++ {
		b.WriteString("  ")
		b.WriteString(m.lines[i])
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.lines) > m.height {
		progress := fmt.Sprintf("  [%d/%d]", m.offset+1, len(m.lines))
		b.WriteString(styles.Muted.Render(progress))
		b.WriteString("\n")
	}

	// Footer
	footer := "  ↑/k: up  ↓/j: down  g: top  G: bottom  Esc: back"
	b.WriteString(styles.Footer.Render(footer))

	return b.String()
}

// DetailStyles holds styles for the detail view.
type DetailStyles struct {
	Header lipgloss.Style
	Footer lipgloss.Style
	Muted  lipgloss.Style
}

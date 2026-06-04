// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a single line in the diff.
type DiffLine struct {
	Content string
	Type    string // "added", "removed", "unchanged"
}

// DiffModel displays config differences.
type DiffModel struct {
	activeAlias   string
	selectedAlias string
	lines         []DiffLine
	offset        int
	height        int
	active        bool
}

// NewDiffModel creates a DiffModel.
func NewDiffModel() DiffModel {
	return DiffModel{}
}

// Show activates the diff view with the given content.
func (m *DiffModel) Show(activeAlias, selectedAlias, activeContent, selectedContent string, height int) {
	m.activeAlias = activeAlias
	m.selectedAlias = selectedAlias
	m.lines = computeDiff(activeContent, selectedContent)
	m.offset = 0
	m.height = height - 4
	m.active = true
}

// Hide deactivates the diff view.
func (m *DiffModel) Hide() {
	m.active = false
}

// IsActive reports whether diff view is active.
func (m DiffModel) IsActive() bool {
	return m.active
}

// ScrollUp scrolls the diff up one line.
func (m *DiffModel) ScrollUp() {
	if m.offset > 0 {
		m.offset--
	}
}

// ScrollDown scrolls the diff down one line.
func (m *DiffModel) ScrollDown() {
	maxOffset := max(len(m.lines)-m.height, 0)
	if m.offset < maxOffset {
		m.offset++
	}
}

// ScrollToTop scrolls to the beginning.
func (m *DiffModel) ScrollToTop() {
	m.offset = 0
}

// ScrollToBottom scrolls to the end.
func (m *DiffModel) ScrollToBottom() {
	m.offset = max(len(m.lines)-m.height, 0)
}

// computeDiff generates diff lines from two strings using simple line-by-line comparison.
func computeDiff(active, selected string) []DiffLine {
	activeLines := strings.Split(active, "\n")
	selectedLines := strings.Split(selected, "\n")

	var lines []DiffLine

	maxLen := len(activeLines)
	if len(selectedLines) > maxLen {
		maxLen = len(selectedLines)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(activeLines) {
			lines = append(lines, DiffLine{Content: "+ " + selectedLines[i], Type: "added"})
		} else if i >= len(selectedLines) {
			lines = append(lines, DiffLine{Content: "- " + activeLines[i], Type: "removed"})
		} else if activeLines[i] != selectedLines[i] {
			lines = append(lines, DiffLine{Content: "- " + activeLines[i], Type: "removed"})
			lines = append(lines, DiffLine{Content: "+ " + selectedLines[i], Type: "added"})
		} else {
			lines = append(lines, DiffLine{Content: "  " + activeLines[i], Type: "unchanged"})
		}
	}

	return lines
}

// Render returns the diff view string.
func (m DiffModel) Render(styles DiffStyles, width int) string {
	var b strings.Builder

	header := fmt.Sprintf("  Diff: %s vs %s", m.activeAlias, m.selectedAlias)
	b.WriteString(styles.Header.Render(header))
	b.WriteString("\n\n")

	end := min(m.offset+m.height, len(m.lines))
	if end <= m.offset && len(m.lines) > 0 {
		end = len(m.lines)
	}

	for i := m.offset; i < end; i++ {
		line := m.lines[i]
		var style lipgloss.Style
		switch line.Type {
		case "added":
			style = styles.Added
		case "removed":
			style = styles.Removed
		default:
			style = styles.Unchanged
		}
		b.WriteString(style.Render(line.Content))
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

// DiffStyles holds styles for the diff view.
type DiffStyles struct {
	Header    lipgloss.Style
	Added     lipgloss.Style
	Removed   lipgloss.Style
	Unchanged lipgloss.Style
	Footer    lipgloss.Style
	Muted     lipgloss.Style
}

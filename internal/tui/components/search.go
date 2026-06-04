package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type SearchModel struct {
	input    textinput.Model
	items    []ListItem
	filtered []ListItem
	cursor   int
	active   bool
	width    int
}

type SearchStyles struct {
	Selected lipgloss.Style
	Normal   lipgloss.Style
}

func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.Focus()
	return SearchModel{
		input: ti,
	}
}

func (m *SearchModel) Activate(items []ListItem) {
	m.active = true
	m.items = items
	m.filtered = items
	m.cursor = 0
	m.input.SetValue("")
	m.input.Focus()
}

func (m *SearchModel) Deactivate() {
	m.active = false
	m.input.SetValue("")
	m.input.Blur()
}

func (m SearchModel) IsActive() bool {
	return m.active
}

func (m *SearchModel) Update(msg tea.KeyMsg) bool {
	if !m.active {
		return false
	}

	switch msg.String() {
	case "esc":
		m.Deactivate()
		return true
	case "enter":
		return true
	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
		}
		return true
	case "down", "ctrl+n":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		return true
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	_ = cmd

	m.filter()
	return true
}

func (m *SearchModel) filter() {
	query := strings.ToLower(strings.TrimSpace(m.input.Value()))
	if query == "" {
		m.filtered = m.items
		m.cursor = 0
		return
	}

	m.filtered = nil
	for _, item := range m.items {
		if strings.HasPrefix(item.Alias, "__group__:") {
			continue
		}
		if strings.Contains(strings.ToLower(item.Alias), query) {
			m.filtered = append(m.filtered, item)
		}
	}
	m.cursor = 0
}

func (m SearchModel) Selected() *ListItem {
	if len(m.filtered) == 0 {
		return nil
	}
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &m.filtered[m.cursor]
}

func (m SearchModel) Render(styles SearchStyles, width int) string {
	var b strings.Builder

	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString("  No matching configs")
	} else {
		for i, item := range m.filtered {
			cursor := "  "
			if i == m.cursor {
				cursor = "❯ "
			}
			line := fmt.Sprintf("%s%-20s → %s", cursor, item.Alias, item.FileName)
			if i == m.cursor {
				b.WriteString(styles.Selected.Render(line))
			} else {
				b.WriteString(styles.Normal.Render(line))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

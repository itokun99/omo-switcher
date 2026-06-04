// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/itokun99/omo-switch/internal/domain"
)

// ListItem represents a single config entry in the list.
type ListItem struct {
	Alias    string
	FileName string
	Group    string
	IsValid  bool
	IsActive bool
}

// ListStyles holds the styles needed by ListModel.
type ListStyles struct {
	GroupHeader lipgloss.Style
	Active      lipgloss.Style
	Inactive    lipgloss.Style
	Cursor      lipgloss.Style
}

// ListModel handles the grouped config list display and navigation.
type ListModel struct {
	items  []ListItem
	cursor int
	offset int
	height int
	active string
}

// NewListModel creates a ListModel from domain groups.
func NewListModel(groups []domain.Group, active string, height int) ListModel {
	items := flattenGroups(groups, active)
	m := ListModel{
		items:  items,
		active: active,
		height: height,
	}
	m.MoveToTop()
	return m
}

func flattenGroups(groups []domain.Group, active string) []ListItem {
	var items []ListItem
	for _, g := range groups {
		if len(g.Configs) == 0 {
			continue
		}
		items = append(items, ListItem{Alias: "__group__:" + g.Name, Group: g.Name})
		for _, cfg := range g.Configs {
			items = append(items, ListItem{
				Alias:    cfg.Alias,
				FileName: cfg.FileName,
				Group:    g.Name,
				IsValid:  cfg.IsValid,
				IsActive: cfg.Alias == active,
			})
		}
	}
	return items
}

// MoveUp moves cursor up, skipping group headers.
func (m *ListModel) MoveUp() {
	for m.cursor > 0 {
		m.cursor--
		if !m.isGroupHeader(m.cursor) {
			m.adjustOffset()
			return
		}
	}
	m.MoveToTop()
}

// MoveDown moves cursor down, skipping group headers.
func (m *ListModel) MoveDown() {
	for m.cursor < len(m.items)-1 {
		m.cursor++
		if !m.isGroupHeader(m.cursor) {
			break
		}
	}
	m.adjustOffset()
}

// MoveToTop moves cursor to first non-header item.
func (m *ListModel) MoveToTop() {
	m.cursor = 0
	for m.cursor < len(m.items) && m.isGroupHeader(m.cursor) {
		m.cursor++
	}
	m.offset = 0
}

// MoveToBottom moves cursor to last non-header item.
func (m *ListModel) MoveToBottom() {
	m.cursor = len(m.items) - 1
	for m.cursor >= 0 && m.isGroupHeader(m.cursor) {
		m.cursor--
	}
	m.adjustOffset()
}

// Items returns all items in the list (including group headers).
func (m ListModel) Items() []ListItem {
	return m.items
}

// SetCursorToItem moves the cursor to the item with the given alias.
func (m *ListModel) SetCursorToItem(alias string) {
	for i, item := range m.items {
		if item.Alias == alias {
			m.cursor = i
			m.adjustOffset()
			return
		}
	}
}

// Selected returns the currently selected item, or nil if on a header.
func (m ListModel) Selected() *ListItem {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return nil
	}
	item := m.items[m.cursor]
	if m.isGroupHeader(m.cursor) {
		return nil
	}
	return &item
}

func (m *ListModel) adjustOffset() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

func (m ListModel) isGroupHeader(index int) bool {
	if index < 0 || index >= len(m.items) {
		return false
	}
	return strings.HasPrefix(m.items[index].Alias, "__group__:")
}

// Render returns the visible portion of the list as a string.
func (m ListModel) Render(styles ListStyles, width int) string {
	var b strings.Builder

	end := min(m.offset+m.height, len(m.items))

	for i := m.offset; i < end; i++ {
		item := m.items[i]

		if m.isGroupHeader(i) {
			groupName := strings.TrimPrefix(item.Alias, "__group__:")
			b.WriteString(styles.GroupHeader.Render(groupName))
			b.WriteString("\n")
			continue
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "❯ "
		}

		marker := ""
		if item.IsActive {
			marker = " ◀ active"
		}

		line := fmt.Sprintf("%s%-20s → %s%s", cursor, item.Alias, item.FileName, marker)

		if i == m.cursor {
			b.WriteString(styles.Cursor.Render(line))
		} else if item.IsActive {
			b.WriteString(styles.Active.Render(line))
		} else {
			b.WriteString(styles.Inactive.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

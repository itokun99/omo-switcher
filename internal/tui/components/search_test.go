package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func createTestItems() []ListItem {
	return []ListItem{
		{Alias: "__group__:Mono", Group: "Mono"},
		{Alias: "claude", FileName: "omo-claude.json", Group: "Mono"},
		{Alias: "gpt", FileName: "omo-gpt.json", Group: "Mono"},
		{Alias: "__group__:Optimized", Group: "Optimized"},
		{Alias: "optimized-high", FileName: "omo-optimized-high.json", Group: "Optimized"},
		{Alias: "optimized-low", FileName: "omo-optimized-low.json", Group: "Optimized"},
	}
}

func TestSearchActivate(t *testing.T) {
	m := NewSearchModel()
	items := createTestItems()

	m.Activate(items)

	if !m.IsActive() {
		t.Error("expected search to be active")
	}
	if len(m.items) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(m.items))
	}
}

func TestSearchDeactivate(t *testing.T) {
	m := NewSearchModel()
	items := createTestItems()

	m.Activate(items)
	m.Deactivate()

	if m.IsActive() {
		t.Error("expected search to be inactive")
	}
	if m.input.Value() != "" {
		t.Errorf("expected empty input, got %q", m.input.Value())
	}
}

func TestSearchFilter(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "filter by claude",
			query:    "claude",
			expected: []string{"claude"},
		},
		{
			name:     "filter by optimized",
			query:    "optimized",
			expected: []string{"optimized-high", "optimized-low"},
		},
		{
			name:     "case insensitive",
			query:    "Claude",
			expected: []string{"claude"},
		},
		{
			name:     "empty query shows all",
			query:    "",
			expected: []string{"__group__:Mono", "claude", "gpt", "__group__:Optimized", "optimized-high", "optimized-low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSearchModel()
			m.Activate(createTestItems())

			if tt.query != "" {
				for _, ch := range tt.query {
					m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
				}
			}

			if len(m.filtered) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(m.filtered))
			}

			for i, exp := range tt.expected {
				if m.filtered[i].Alias != exp {
					t.Errorf("expected %q at index %d, got %q", exp, i, m.filtered[i].Alias)
				}
			}
		})
	}
}

func TestSearchFilterEmpty(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	for _, ch := range "nonexistent" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if len(m.filtered) != 0 {
		t.Errorf("expected 0 results, got %d", len(m.filtered))
	}
}

func TestSearchNavigation(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	for _, ch := range "optimized" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestSearchSelected(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	for _, ch := range "claude" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected selected item, got nil")
	}
	if sel.Alias != "claude" {
		t.Errorf("expected claude, got %s", sel.Alias)
	}
}

func TestSearchSelectedEmpty(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	for _, ch := range "nonexistent" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if m.Selected() != nil {
		t.Error("expected nil for empty results")
	}
}

func TestSearchEscDeactivates(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	consumed := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !consumed {
		t.Error("expected esc to be consumed")
	}
	if m.IsActive() {
		t.Error("expected search deactivated after esc")
	}
}

func TestSearchInactiveIgnoresKeys(t *testing.T) {
	m := NewSearchModel()

	consumed := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if consumed {
		t.Error("inactive search should not consume keys")
	}
}

func TestSearchSkipsGroupHeaders(t *testing.T) {
	m := NewSearchModel()
	m.Activate(createTestItems())

	for _, ch := range "Mono" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	if len(m.filtered) != 0 {
		t.Errorf("expected 0 results for group header query, got %d", len(m.filtered))
	}
}

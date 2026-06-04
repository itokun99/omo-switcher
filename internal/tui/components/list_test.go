package components

import (
	"testing"

	"github.com/itokun99/omo-switch/internal/domain"
)

func createTestGroups() []domain.Group {
	return []domain.Group{
		{
			Name: "Mono",
			Configs: []domain.Config{
				{Alias: "claude", FileName: "omo-claude.json"},
				{Alias: "gpt", FileName: "omo-gpt.json"},
			},
		},
		{
			Name: "Optimized",
			Configs: []domain.Config{
				{Alias: "optimized-high", FileName: "omo-optimized-high.json"},
			},
		},
	}
}

func TestFlattenGroups(t *testing.T) {
	groups := createTestGroups()
	items := flattenGroups(groups, "claude")

	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}

	if items[0].Alias != "__group__:Mono" {
		t.Errorf("expected group header, got %s", items[0].Alias)
	}

	if items[1].Alias != "claude" || !items[1].IsActive {
		t.Errorf("expected active claude, got %v", items[1])
	}
}

func TestNavigation(t *testing.T) {
	tests := []struct {
		name     string
		active   string
		actions  func(*ListModel)
		expected string
	}{
		{
			name:   "start at first non-header",
			active: "claude",
			actions: func(m *ListModel) {},
			expected: "claude",
		},
		{
			name:   "move down to gpt",
			active: "claude",
			actions: func(m *ListModel) { m.MoveDown() },
			expected: "gpt",
		},
		{
			name:   "move up stays at claude",
			active: "claude",
			actions: func(m *ListModel) { m.MoveUp() },
			expected: "claude",
		},
		{
			name:   "move down then up",
			active: "claude",
			actions: func(m *ListModel) { m.MoveDown(); m.MoveUp() },
			expected: "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := createTestGroups()
			model := NewListModel(groups, tt.active, 10)
			tt.actions(&model)

			selected := model.Selected()
			if selected == nil || selected.Alias != tt.expected {
				t.Errorf("expected %s selected, got %v", tt.expected, selected)
			}
		})
	}
}

func TestMoveToBottom(t *testing.T) {
	groups := createTestGroups()
	model := NewListModel(groups, "claude", 10)

	model.MoveToBottom()
	selected := model.Selected()
	if selected == nil || selected.Alias != "optimized-high" {
		t.Errorf("expected optimized-high at bottom, got %v", selected)
	}
}

func TestMoveToTop(t *testing.T) {
	groups := createTestGroups()
	model := NewListModel(groups, "optimized-high", 10)

	model.MoveToBottom()
	model.MoveToTop()
	selected := model.Selected()
	if selected == nil || selected.Alias != "claude" {
		t.Errorf("expected claude at top, got %v", selected)
	}
}

func TestSelectedOnGroupHeader(t *testing.T) {
	groups := createTestGroups()
	model := NewListModel(groups, "claude", 10)

	model.cursor = 0
	if model.Selected() != nil {
		t.Errorf("expected nil when cursor is on group header")
	}
}

func TestEmptyGroups(t *testing.T) {
	groups := []domain.Group{
		{Name: "Mono", Configs: []domain.Config{}},
	}
	model := NewListModel(groups, "", 10)

	if len(model.items) != 0 {
		t.Errorf("expected 0 items for empty groups, got %d", len(model.items))
	}

	selected := model.Selected()
	if selected != nil {
		t.Errorf("expected nil for empty list, got %v", selected)
	}
}

func TestAdjustOffset(t *testing.T) {
	groups := createTestGroups()
	model := NewListModel(groups, "claude", 2)

	model.MoveDown()
	model.MoveDown()
	model.MoveDown()

	if model.cursor < model.offset {
		t.Errorf("cursor %d should be >= offset %d", model.cursor, model.offset)
	}
	if model.cursor >= model.offset+model.height {
		t.Errorf("cursor %d should be < offset+height %d", model.cursor, model.offset+model.height)
	}
}

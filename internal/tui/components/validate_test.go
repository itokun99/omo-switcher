package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testValidateStyles() ValidateStyles {
	return ValidateStyles{
		Valid:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		Invalid: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

func TestValidateShow(t *testing.T) {
	m := NewValidateModel()

	results := []ValidationResult{
		{Alias: "claude", IsValid: true},
		{Alias: "broken", IsValid: false, Error: "missing agents key"},
	}

	m.Show(results, 24)

	if !m.IsActive() {
		t.Error("expected IsActive() = true after Show")
	}
	if len(m.results) != 2 {
		t.Errorf("expected 2 results, got %d", len(m.results))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestValidateHide(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{{Alias: "test", IsValid: true}}, 24)

	if !m.IsActive() {
		t.Fatal("expected active before hide")
	}

	m.Hide()
	if m.IsActive() {
		t.Error("expected IsActive() = false after Hide")
	}
}

func TestValidateNavigation(t *testing.T) {
	results := []ValidationResult{
		{Alias: "a", IsValid: true},
		{Alias: "b", IsValid: true},
		{Alias: "c", IsValid: false, Error: "invalid"},
	}

	tests := []struct {
		name     string
		moves    func(m *ValidateModel)
		expected int
	}{
		{
			name:     "initial cursor at 0",
			moves:    func(m *ValidateModel) {},
			expected: 0,
		},
		{
			name: "move down once",
			moves: func(m *ValidateModel) {
				m.MoveDown()
			},
			expected: 1,
		},
		{
			name: "move down past end stays at last",
			moves: func(m *ValidateModel) {
				for range 10 {
					m.MoveDown()
				}
			},
			expected: 2,
		},
		{
			name: "move up from bottom",
			moves: func(m *ValidateModel) {
				for range 10 {
					m.MoveDown()
				}
				m.MoveUp()
			},
			expected: 1,
		},
		{
			name: "move up past 0 stays at 0",
			moves: func(m *ValidateModel) {
				m.MoveUp()
				m.MoveUp()
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewValidateModel()
			m.Show(results, 24)
			tt.moves(&m)
			if m.cursor != tt.expected {
				t.Errorf("expected cursor %d, got %d", tt.expected, m.cursor)
			}
		})
	}
}

func TestValidateRenderEmpty(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{}, 24)

	styles := testValidateStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "No configs to validate") {
		t.Error("expected 'No configs to validate' in output")
	}
}

func TestValidateRenderValid(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{
		{Alias: "claude", IsValid: true},
	}, 24)

	styles := testValidateStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "Config Validation") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "claude") {
		t.Error("expected alias in output")
	}
	if !strings.Contains(output, "✓") {
		t.Error("expected valid checkmark")
	}
}

func TestValidateRenderInvalid(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{
		{Alias: "broken", IsValid: false, Error: "missing agents key"},
	}, 24)

	styles := testValidateStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "✗") {
		t.Error("expected invalid cross mark")
	}
	if !strings.Contains(output, "broken") {
		t.Error("expected alias in output")
	}
	if !strings.Contains(output, "missing agents key") {
		t.Error("expected error message in output")
	}
}

func TestValidateRenderMixed(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{
		{Alias: "good", IsValid: true},
		{Alias: "bad", IsValid: false, Error: "parse error"},
	}, 24)

	styles := testValidateStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "✓") {
		t.Error("expected valid checkmark for good config")
	}
	if !strings.Contains(output, "✗") {
		t.Error("expected invalid cross for bad config")
	}
	if !strings.Contains(output, "parse error") {
		t.Error("expected error message")
	}
}

func TestValidateRenderCursorIndicator(t *testing.T) {
	m := NewValidateModel()
	m.Show([]ValidationResult{
		{Alias: "first", IsValid: true},
		{Alias: "second", IsValid: true},
	}, 24)

	styles := testValidateStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "❯") {
		t.Error("expected cursor indicator")
	}
}

package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHelpToggle(t *testing.T) {
	m := NewHelpModel()

	if m.IsActive() {
		t.Error("expected inactive after creation")
	}

	m.Toggle()
	if !m.IsActive() {
		t.Error("expected active after toggle")
	}

	m.Toggle()
	if m.IsActive() {
		t.Error("expected inactive after second toggle")
	}
}

func TestHelpHide(t *testing.T) {
	m := NewHelpModel()
	m.Toggle()

	if !m.IsActive() {
		t.Fatal("expected active before hide")
	}

	m.Hide()
	if m.IsActive() {
		t.Error("expected inactive after hide")
	}
}

func TestHelpRenderInactive(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(80, 24)

	rendered := m.Render(HelpStyles{
		Box: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
	})

	if rendered != "" {
		t.Errorf("expected empty string when inactive, got %q", rendered)
	}
}

func TestHelpRender(t *testing.T) {
	m := NewHelpModel()
	m.Toggle()
	m.SetSize(80, 24)

	rendered := m.Render(HelpStyles{
		Box: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
	})

	if rendered == "" {
		t.Fatal("expected non-empty output when active")
	}

	expected := []string{
		"Key Bindings",
		"Navigation",
		"Actions",
		"Utilities",
		"General",
		"Move up",
		"Move down",
		"Search configs",
		"Toggle this help",
		"Quit",
	}

	for _, want := range expected {
		if !strings.Contains(rendered, want) {
			t.Errorf("expected %q in rendered output", want)
		}
	}
}

func TestHelpSetSize(t *testing.T) {
	m := NewHelpModel()
	m.SetSize(100, 30)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 30 {
		t.Errorf("expected height 30, got %d", m.height)
	}
}

func TestHelpRenderSmallDimensions(t *testing.T) {
	m := NewHelpModel()
	m.Toggle()
	m.SetSize(20, 10)

	rendered := m.Render(HelpStyles{
		Box: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
	})

	if rendered == "" {
		t.Fatal("expected non-empty output even with small dimensions")
	}

	if !strings.Contains(rendered, "Key Bindings") {
		t.Error("expected key bindings content")
	}
}

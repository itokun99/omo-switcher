package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusSetActive(t *testing.T) {
	m := NewStatusModel()
	m.SetActive("claude", "Mono")

	if m.activeAlias != "claude" {
		t.Errorf("expected claude, got %s", m.activeAlias)
	}
	if m.activeGroup != "Mono" {
		t.Errorf("expected Mono, got %s", m.activeGroup)
	}
}

func TestStatusMessage(t *testing.T) {
	m := NewStatusModel()
	m.SetActive("claude", "Mono")
	m.SetMessage("Switched successfully!")

	rendered := m.Render(StatusStyles{
		Bar: lipgloss.NewStyle(),
	})
	if !strings.Contains(rendered, "Switched successfully!") {
		t.Error("expected message in rendered output")
	}

	m.ClearMessage()
	rendered = m.Render(StatusStyles{
		Bar: lipgloss.NewStyle(),
	})
	if !strings.Contains(rendered, "claude") {
		t.Error("expected claude after clearing message")
	}
}

func TestStatusNoActive(t *testing.T) {
	m := NewStatusModel()
	rendered := m.Render(StatusStyles{
		Bar: lipgloss.NewStyle(),
	})
	if !strings.Contains(rendered, "No active config") {
		t.Error("expected 'No active config' when none set")
	}
}

func TestStatusWidthPadding(t *testing.T) {
	m := NewStatusModel()
	m.SetActive("claude", "Mono")
	m.SetWidth(80)

	rendered := m.Render(StatusStyles{
		Bar: lipgloss.NewStyle(),
	})

	if !strings.Contains(rendered, "claude") {
		t.Error("expected claude in rendered output")
	}
	if !strings.Contains(rendered, "navigate") {
		t.Error("expected key hints in rendered output")
	}
}

func TestStatusMessageOverridesActive(t *testing.T) {
	m := NewStatusModel()
	m.SetActive("claude", "Mono")
	m.SetMessage("Error: invalid config")

	rendered := m.Render(StatusStyles{
		Bar: lipgloss.NewStyle(),
	})

	if strings.Contains(rendered, "claude") {
		t.Error("message should override active config display")
	}
	if !strings.Contains(rendered, "Error: invalid config") {
		t.Error("expected error message in rendered output")
	}
}

func TestStatusSetWidth(t *testing.T) {
	m := NewStatusModel()
	m.SetWidth(100)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
}

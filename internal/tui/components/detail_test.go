package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testDetailStyles() DetailStyles {
	return DetailStyles{
		Header: lipgloss.NewStyle().Bold(true),
		Footer: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

const testJSON = `{
  "agents": {
    "sisyphus": {
      "model": "opencode-go/kimi-k2.6"
    }
  },
  "categories": {
    "deep": {
      "model": "opencode-go/deepseek-v4-pro"
    }
  }
}`

func TestDetailShow(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 20)

	if !m.IsActive() {
		t.Error("expected IsActive() = true after Show")
	}
	if m.alias != "claude" {
		t.Errorf("expected alias claude, got %s", m.alias)
	}
	if m.filename != "omo-claude.json" {
		t.Errorf("expected filename omo-claude.json, got %s", m.filename)
	}
	if len(m.lines) != strings.Count(testJSON, "\n")+1 {
		t.Errorf("expected %d lines, got %d", strings.Count(testJSON, "\n")+1, len(m.lines))
	}
	if m.offset != 0 {
		t.Errorf("expected offset 0, got %d", m.offset)
	}
}

func TestDetailHide(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 20)
	m.Hide()

	if m.IsActive() {
		t.Error("expected IsActive() = false after Hide")
	}
}

func TestDetailScrollDown(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 5)

	m.ScrollDown()
	if m.offset != 1 {
		t.Errorf("expected offset 1, got %d", m.offset)
	}

	m.ScrollDown()
	if m.offset != 2 {
		t.Errorf("expected offset 2, got %d", m.offset)
	}
}

func TestDetailScrollUp(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 5)

	m.ScrollDown()
	m.ScrollDown()
	m.ScrollUp()

	if m.offset != 1 {
		t.Errorf("expected offset 1, got %d", m.offset)
	}
}

func TestDetailScrollBoundary(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 5)

	// Scroll past end should stop
	for range 100 {
		m.ScrollDown()
	}

	maxOffset := max(len(m.lines)-m.height, 0)
	if m.offset != maxOffset {
		t.Errorf("expected offset %d (max), got %d", maxOffset, m.offset)
	}

	// Scroll up from 0 should stay at 0
	m2 := NewDetailModel()
	m2.Show("claude", "omo-claude.json", testJSON, 20)
	m2.ScrollUp()
	if m2.offset != 0 {
		t.Errorf("expected offset 0, got %d", m2.offset)
	}
}

func TestDetailRender(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 20)

	styles := testDetailStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "claude") {
		t.Error("rendered output should contain alias")
	}
	if !strings.Contains(output, "omo-claude.json") {
		t.Error("rendered output should contain filename")
	}
	if !strings.Contains(output, "Esc") {
		t.Error("rendered output should contain footer with Esc")
	}
}

func TestDetailRenderSmallHeight(t *testing.T) {
	m := NewDetailModel()
	m.Show("claude", "omo-claude.json", testJSON, 6)

	styles := testDetailStyles()
	output := m.Render(styles, 80)

	// Should not panic, should render something
	if output == "" {
		t.Error("rendered output should not be empty")
	}
}



package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testDiffStyles() DiffStyles {
	return DiffStyles{
		Header:    lipgloss.NewStyle().Bold(true),
		Added:     lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		Removed:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Unchanged: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Footer:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

const testActiveContent = `{
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

const testSelectedContent = `{
  "agents": {
    "sisyphus": {
      "model": "opencode-go/gpt-4o"
    }
  },
  "categories": {
    "deep": {
      "model": "opencode-go/deepseek-v4-pro"
    },
    "fast": {
      "model": "opencode-go/claude-sonnet-4"
    }
  }
}`

func TestDiffShow(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 20)

	if !m.IsActive() {
		t.Error("expected IsActive() = true after Show")
	}
	if m.activeAlias != "optimized-high" {
		t.Errorf("expected activeAlias optimized-high, got %s", m.activeAlias)
	}
	if m.selectedAlias != "gpt" {
		t.Errorf("expected selectedAlias gpt, got %s", m.selectedAlias)
	}
	if m.offset != 0 {
		t.Errorf("expected offset 0, got %d", m.offset)
	}
	if len(m.lines) == 0 {
		t.Error("expected non-empty diff lines")
	}
}

func TestDiffHide(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 20)
	m.Hide()

	if m.IsActive() {
		t.Error("expected IsActive() = false after Hide")
	}
}

func TestDiffScrollDown(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 5)

	m.ScrollDown()
	if m.offset != 1 {
		t.Errorf("expected offset 1, got %d", m.offset)
	}

	m.ScrollDown()
	if m.offset != 2 {
		t.Errorf("expected offset 2, got %d", m.offset)
	}
}

func TestDiffScrollUp(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 5)

	m.ScrollDown()
	m.ScrollDown()
	m.ScrollUp()

	if m.offset != 1 {
		t.Errorf("expected offset 1, got %d", m.offset)
	}
}

func TestDiffScrollBoundary(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 5)

	for range 100 {
		m.ScrollDown()
	}

	maxOffset := max(len(m.lines)-m.height, 0)
	if m.offset != maxOffset {
		t.Errorf("expected offset %d (max), got %d", maxOffset, m.offset)
	}

	m2 := NewDiffModel()
	m2.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 50)
	m2.ScrollUp()
	if m2.offset != 0 {
		t.Errorf("expected offset 0, got %d", m2.offset)
	}
}

func TestDiffScrollToTop(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 5)

	m.ScrollDown()
	m.ScrollDown()
	m.ScrollDown()
	m.ScrollToTop()

	if m.offset != 0 {
		t.Errorf("expected offset 0 after ScrollToTop, got %d", m.offset)
	}
}

func TestDiffScrollToBottom(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 5)

	m.ScrollToBottom()

	expected := max(len(m.lines)-m.height, 0)
	if m.offset != expected {
		t.Errorf("expected offset %d after ScrollToBottom, got %d", expected, m.offset)
	}
}

func TestDiffRender(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 20)

	styles := testDiffStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "Diff:") {
		t.Error("rendered output should contain 'Diff:' header")
	}
	if !strings.Contains(output, "optimized-high") {
		t.Error("rendered output should contain active alias")
	}
	if !strings.Contains(output, "gpt") {
		t.Error("rendered output should contain selected alias")
	}
	if !strings.Contains(output, "Esc") {
		t.Error("rendered output should contain footer with Esc")
	}
}

func TestDiffRenderSmallHeight(t *testing.T) {
	m := NewDiffModel()
	m.Show("optimized-high", "gpt", testActiveContent, testSelectedContent, 6)

	styles := testDiffStyles()
	output := m.Render(styles, 80)

	if output == "" {
		t.Error("rendered output should not be empty")
	}
}

func TestComputeDiff(t *testing.T) {
	tests := []struct {
		name     string
		active   string
		selected string
		want     []DiffLine
	}{
		{
			name:     "identical content",
			active:   "line1\nline2\nline3",
			selected: "line1\nline2\nline3",
			want: []DiffLine{
				{Content: "  line1", Type: "unchanged"},
				{Content: "  line2", Type: "unchanged"},
				{Content: "  line3", Type: "unchanged"},
			},
		},
		{
			name:     "added lines",
			active:   "line1\nline2",
			selected: "line1\nline2\nline3",
			want: []DiffLine{
				{Content: "  line1", Type: "unchanged"},
				{Content: "  line2", Type: "unchanged"},
				{Content: "+ line3", Type: "added"},
			},
		},
		{
			name:     "removed lines",
			active:   "line1\nline2\nline3",
			selected: "line1\nline2",
			want: []DiffLine{
				{Content: "  line1", Type: "unchanged"},
				{Content: "  line2", Type: "unchanged"},
				{Content: "- line3", Type: "removed"},
			},
		},
		{
			name:     "changed line",
			active:   "line1\nold-value\nline3",
			selected: "line1\nnew-value\nline3",
			want: []DiffLine{
				{Content: "  line1", Type: "unchanged"},
				{Content: "- old-value", Type: "removed"},
				{Content: "+ new-value", Type: "added"},
				{Content: "  line3", Type: "unchanged"},
			},
		},
		{
			name:     "empty active",
			active:   "",
			selected: "line1\nline2",
			want: []DiffLine{
				{Content: "- ", Type: "removed"},
				{Content: "+ line1", Type: "added"},
				{Content: "+ line2", Type: "added"},
			},
		},
		{
			name:     "empty selected",
			active:   "line1\nline2",
			selected: "",
			want: []DiffLine{
				{Content: "- line1", Type: "removed"},
				{Content: "+ ", Type: "added"},
				{Content: "- line2", Type: "removed"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeDiff(tt.active, tt.selected)

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d diff lines, got %d", len(tt.want), len(got))
			}

			for i, wantLine := range tt.want {
				if got[i].Content != wantLine.Content {
					t.Errorf("line %d: expected content %q, got %q", i, wantLine.Content, got[i].Content)
				}
				if got[i].Type != wantLine.Type {
					t.Errorf("line %d: expected type %q, got %q", i, wantLine.Type, got[i].Type)
				}
			}
		})
	}
}

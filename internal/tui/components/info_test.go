package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testInfoStyles() InfoStyles {
	return InfoStyles{
		Label: lipgloss.NewStyle().Foreground(lipgloss.Color("60")),
	}
}

func TestInfoShow(t *testing.T) {
	m := NewInfoModel()
	info := ConfigInfo{
		Alias:      "claude",
		FileName:   "omo-claude.json",
		FilePath:   "/home/user/.config/opencode/omo_configs/omo-claude.json",
		FileSize:   2048,
		ModifiedAt: "2025-01-15 10:30:00",
		IsValid:    true,
	}

	m.Show(info)

	if !m.IsActive() {
		t.Error("expected IsActive() = true after Show")
	}
	if m.info.Alias != "claude" {
		t.Errorf("expected alias claude, got %s", m.info.Alias)
	}
	if m.info.FileSize != 2048 {
		t.Errorf("expected FileSize 2048, got %d", m.info.FileSize)
	}
}

func TestInfoHide(t *testing.T) {
	m := NewInfoModel()
	info := ConfigInfo{Alias: "test"}
	m.Show(info)
	m.Hide()

	if m.IsActive() {
		t.Error("expected IsActive() = false after Hide")
	}
}

func TestInfoRender(t *testing.T) {
	m := NewInfoModel()
	m.Show(ConfigInfo{
		Alias:      "claude",
		FileName:   "omo-claude.json",
		FilePath:   "/home/user/.config/opencode/omo_configs/omo-claude.json",
		FileSize:   1536,
		ModifiedAt: "2025-01-15 10:30:00",
	})

	styles := testInfoStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "Config Info") {
		t.Error("rendered output should contain header")
	}
	if !strings.Contains(output, "claude") {
		t.Error("rendered output should contain alias")
	}
	if !strings.Contains(output, "omo-claude.json") {
		t.Error("rendered output should contain filename")
	}
	if !strings.Contains(output, "1.5 KB") {
		t.Error("rendered output should contain formatted size")
	}
}

func TestInfoRenderEmptyValue(t *testing.T) {
	m := NewInfoModel()
	m.Show(ConfigInfo{Alias: "test"})

	styles := testInfoStyles()
	output := m.Render(styles, 80)

	if !strings.Contains(output, "(unknown)") {
		t.Error("rendered output should show (unknown) for empty values")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes under 1KB", 512, "512 B"},
		{"exactly 1KB", 1024, "1.0 KB"},
		{"fractional KB", 1536, "1.5 KB"},
		{"large KB", 5120, "5.0 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestGetConfigInfo(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "omo-test.json")
	content := `{"agents": {"sisyphus": {"model": "test"}}}`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	info := GetConfigInfo("test", filePath)

	if info.Alias != "test" {
		t.Errorf("expected alias test, got %s", info.Alias)
	}
	if info.FileName != "omo-test.json" {
		t.Errorf("expected filename omo-test.json, got %s", info.FileName)
	}
	if info.FilePath != filePath {
		t.Errorf("expected FilePath %s, got %s", filePath, info.FilePath)
	}
	if info.FileSize == 0 {
		t.Error("expected non-zero FileSize")
	}
	if info.ModifiedAt == "" {
		t.Error("expected non-empty ModifiedAt")
	}
}

func TestGetConfigInfoMissingFile(t *testing.T) {
	info := GetConfigInfo("missing", "/nonexistent/path/omo-missing.json")

	if info.Alias != "missing" {
		t.Errorf("expected alias missing, got %s", info.Alias)
	}
	if info.FileSize != 0 {
		t.Errorf("expected FileSize 0 for missing file, got %d", info.FileSize)
	}
	if info.ModifiedAt != "" {
		t.Errorf("expected empty ModifiedAt for missing file, got %s", info.ModifiedAt)
	}
}

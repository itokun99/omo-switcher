package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

func testBackups() []infrastructure.BackupInfo {
	return []infrastructure.BackupInfo{
		{Timestamp: "2025-01-02T15-04-05-000Z", FileName: "oh-my-openagent.2025-01-02T15-04-05-000Z.json"},
		{Timestamp: "2025-01-01T10-00-00-000Z", FileName: "oh-my-openagent.2025-01-01T10-00-00-000Z.json"},
		{Timestamp: "2024-12-31T08-30-00-000Z", FileName: "oh-my-openagent.2024-12-31T08-30-00-000Z.json"},
	}
}

func TestBackupShow(t *testing.T) {
	model := NewBackupModel()
	backups := testBackups()

	model.Show(backups, 24)

	if !model.IsActive() {
		t.Error("expected IsActive() true after Show")
	}
	if len(model.backups) != 3 {
		t.Errorf("expected 3 backups, got %d", len(model.backups))
	}
	if model.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.cursor)
	}
}

func TestBackupHide(t *testing.T) {
	model := NewBackupModel()
	model.Show(testBackups(), 24)
	model.Hide()

	if model.IsActive() {
		t.Error("expected IsActive() false after Hide")
	}
	if model.IsConfirming() {
		t.Error("expected IsConfirming() false after Hide")
	}
}

func TestBackupNavigation(t *testing.T) {
	tests := []struct {
		name     string
		actions  func(*BackupModel)
		expected int
	}{
		{
			name:     "initial position",
			actions:  func(m *BackupModel) {},
			expected: 0,
		},
		{
			name:     "move down once",
			actions:  func(m *BackupModel) { m.MoveDown() },
			expected: 1,
		},
		{
			name:     "move down twice",
			actions:  func(m *BackupModel) { m.MoveDown(); m.MoveDown() },
			expected: 2,
		},
		{
			name:     "move down at bottom stays",
			actions:  func(m *BackupModel) { m.MoveDown(); m.MoveDown(); m.MoveDown() },
			expected: 2,
		},
		{
			name:     "move up at top stays",
			actions:  func(m *BackupModel) { m.MoveUp() },
			expected: 0,
		},
		{
			name:     "move down then up",
			actions:  func(m *BackupModel) { m.MoveDown(); m.MoveUp() },
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewBackupModel()
			model.Show(testBackups(), 24)
			tt.actions(&model)

			if model.cursor != tt.expected {
				t.Errorf("expected cursor %d, got %d", tt.expected, model.cursor)
			}
		})
	}
}

func TestBackupConfirm(t *testing.T) {
	model := NewBackupModel()
	model.Show(testBackups(), 24)

	if model.IsConfirming() {
		t.Error("expected not confirming initially")
	}

	model.StartConfirm()
	if !model.IsConfirming() {
		t.Error("expected confirming after StartConfirm")
	}

	model.CancelConfirm()
	if model.IsConfirming() {
		t.Error("expected not confirming after CancelConfirm")
	}
}

func TestBackupSelected(t *testing.T) {
	model := NewBackupModel()
	model.Show(testBackups(), 24)

	selected := model.Selected()
	if selected != "2025-01-02T15-04-05-000Z" {
		t.Errorf("expected first backup timestamp, got %q", selected)
	}

	model.MoveDown()
	selected = model.Selected()
	if selected != "2025-01-01T10-00-00-000Z" {
		t.Errorf("expected second backup timestamp, got %q", selected)
	}
}

func TestBackupSelectedEmpty(t *testing.T) {
	model := NewBackupModel()
	model.Show([]infrastructure.BackupInfo{}, 24)

	if model.Selected() != "" {
		t.Errorf("expected empty string for no backups, got %q", model.Selected())
	}
}

func TestBackupStartConfirmEmpty(t *testing.T) {
	model := NewBackupModel()
	model.Show([]infrastructure.BackupInfo{}, 24)
	model.StartConfirm()

	if model.IsConfirming() {
		t.Error("expected not confirming when no backups exist")
	}
}

func TestBackupRender(t *testing.T) {
	model := NewBackupModel()
	model.Show(testBackups(), 24)

	styles := BackupStyles{
		Selected: lipgloss.NewStyle().Bold(true),
		Normal:   lipgloss.NewStyle(),
		Confirm:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}

	output := model.Render(styles, 80)

	if !strings.Contains(output, "Backup Manager") {
		t.Error("expected 'Backup Manager' in output")
	}
	if !strings.Contains(output, "oh-my-openagent") {
		t.Error("expected backup filename in output")
	}
}

func TestBackupRenderEmpty(t *testing.T) {
	model := NewBackupModel()
	model.Show([]infrastructure.BackupInfo{}, 24)

	styles := BackupStyles{
		Selected: lipgloss.NewStyle(),
		Normal:   lipgloss.NewStyle(),
		Confirm:  lipgloss.NewStyle(),
	}

	output := model.Render(styles, 80)

	if !strings.Contains(output, "No backups found") {
		t.Error("expected 'No backups found' in output")
	}
}

func TestBackupRenderConfirm(t *testing.T) {
	model := NewBackupModel()
	model.Show(testBackups(), 24)
	model.StartConfirm()

	styles := BackupStyles{
		Selected: lipgloss.NewStyle(),
		Normal:   lipgloss.NewStyle(),
		Confirm:  lipgloss.NewStyle(),
	}

	output := model.Render(styles, 80)

	if !strings.Contains(output, "Restore this backup?") {
		t.Error("expected restore prompt in output")
	}
}

package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

// BackupModel displays and manages backups.
type BackupModel struct {
	backups []infrastructure.BackupInfo
	cursor  int
	active  bool
	height  int
	confirm bool // waiting for restore confirmation
}

// NewBackupModel creates a BackupModel.
func NewBackupModel() BackupModel {
	return BackupModel{}
}

// Show activates the backup manager with the given backups.
func (m *BackupModel) Show(backups []infrastructure.BackupInfo, height int) {
	m.backups = backups
	m.cursor = 0
	m.height = height - 4
	m.active = true
	m.confirm = false
}

// Hide deactivates the backup manager.
func (m *BackupModel) Hide() {
	m.active = false
	m.confirm = false
}

// IsActive reports whether backup manager is active.
func (m BackupModel) IsActive() bool {
	return m.active
}

// IsConfirming reports whether waiting for confirmation.
func (m BackupModel) IsConfirming() bool {
	return m.confirm
}

// MoveUp moves cursor up.
func (m *BackupModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveDown moves cursor down.
func (m *BackupModel) MoveDown() {
	if m.cursor < len(m.backups)-1 {
		m.cursor++
	}
}

// Selected returns the selected backup timestamp, or empty string.
func (m BackupModel) Selected() string {
	if m.cursor < 0 || m.cursor >= len(m.backups) {
		return ""
	}
	return m.backups[m.cursor].Timestamp
}

// StartConfirm enters confirmation mode for restore.
func (m *BackupModel) StartConfirm() {
	if len(m.backups) > 0 {
		m.confirm = true
	}
}

// CancelConfirm cancels confirmation mode.
func (m *BackupModel) CancelConfirm() {
	m.confirm = false
}

// Render returns the backup manager view string.
func (m BackupModel) Render(styles BackupStyles, width int) string {
	var b strings.Builder

	b.WriteString("  Backup Manager\n\n")

	if len(m.backups) == 0 {
		b.WriteString("  No backups found")
		return b.String()
	}

	for i, backup := range m.backups {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s", cursor, backup.FileName)
		if i == m.cursor {
			b.WriteString(styles.Selected.Render(line))
		} else {
			b.WriteString(styles.Normal.Render(line))
		}
		b.WriteString("\n")
	}

	if m.confirm {
		b.WriteString("\n")
		b.WriteString(styles.Confirm.Render("  Restore this backup? (y/n)"))
	}

	return b.String()
}

// BackupStyles holds styles for the backup manager.
type BackupStyles struct {
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Confirm  lipgloss.Style
}

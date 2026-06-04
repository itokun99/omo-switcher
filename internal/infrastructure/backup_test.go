package infrastructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesystemBackupManager_CreateBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "oh-my-openagent.json")
	writeTestFile(t, target, `{"original": "config"}`)

	backupDir := filepath.Join(dir, "backups")
	mgr := NewFilesystemBackupManagerWithPath(backupDir, target)

	path, err := mgr.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if !strings.HasPrefix(path, backupDir) {
		t.Errorf("backup path %q not in backup dir %q", path, backupDir)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(got) != `{"original": "config"}` {
		t.Errorf("backup content = %q, want %q", string(got), `{"original": "config"}`)
	}
}

func TestFilesystemBackupManager_CreateBackup_NoTarget(t *testing.T) {
	dir := t.TempDir()
	mgr := NewFilesystemBackupManagerWithPath(filepath.Join(dir, "backups"), filepath.Join(dir, "missing.json"))

	_, err := mgr.CreateBackup()
	if err == nil {
		t.Fatal("CreateBackup() expected error when target missing")
	}
}

func TestFilesystemBackupManager_ListBackups(t *testing.T) {
	dir := t.TempDir()
	backupDir := filepath.Join(dir, "backups")
	os.MkdirAll(backupDir, 0o755)

	names := []string{
		"oh-my-openagent.2024-01-15T10-30-00-000Z.json",
		"oh-my-openagent.2024-01-16T12-00-00-000Z.json",
		"oh-my-openagent.2024-01-14T08-00-00-000Z.json",
		"unrelated.txt",
	}
	for _, n := range names {
		writeTestFile(t, filepath.Join(backupDir, n), "{}")
	}

	mgr := NewFilesystemBackupManagerWithPath(backupDir, filepath.Join(dir, "target.json"))
	backups, err := mgr.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if len(backups) != 3 {
		t.Fatalf("ListBackups() returned %d backups, want 3", len(backups))
	}

	if backups[0].Timestamp < backups[1].Timestamp {
		t.Errorf("backups not sorted newest first: %s < %s", backups[0].Timestamp, backups[1].Timestamp)
	}
}

func TestFilesystemBackupManager_ListBackups_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	mgr := NewFilesystemBackupManagerWithPath(filepath.Join(dir, "backups"), filepath.Join(dir, "target.json"))

	backups, err := mgr.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("ListBackups() = %d backups, want 0", len(backups))
	}
}

func TestFilesystemBackupManager_RestoreBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "oh-my-openagent.json")
	backupDir := filepath.Join(dir, "backups")

	writeTestFile(t, target, `{"old": "config"}`)
	os.MkdirAll(backupDir, 0o755)

	backupContent := `{"restored": "config"}`
	backupFile := filepath.Join(backupDir, "oh-my-openagent.2024-01-15T10-30-00-000Z.json")
	writeTestFile(t, backupFile, backupContent)

	mgr := NewFilesystemBackupManagerWithPath(backupDir, target)
	if err := mgr.RestoreBackup("2024-01-15T10-30-00-000Z"); err != nil {
		t.Fatalf("RestoreBackup() error = %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading restored file: %v", err)
	}
	if string(got) != backupContent {
		t.Errorf("restored content = %q, want %q", string(got), backupContent)
	}
}

func TestFilesystemBackupManager_RestoreBackup_NotFound(t *testing.T) {
	dir := t.TempDir()
	mgr := NewFilesystemBackupManagerWithPath(filepath.Join(dir, "backups"), filepath.Join(dir, "target.json"))

	err := mgr.RestoreBackup("nonexistent")
	if err == nil {
		t.Fatal("RestoreBackup() expected error for missing timestamp")
	}
}

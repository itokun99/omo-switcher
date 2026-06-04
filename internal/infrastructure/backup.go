package infrastructure

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupManager defines the interface for backup operations.
type BackupManager interface {
	CreateBackup() (string, error)
	ListBackups() ([]BackupInfo, error)
	RestoreBackup(timestamp string) error
}

// BackupInfo holds metadata about a backup.
type BackupInfo struct {
	Timestamp string
	FilePath  string
	FileName  string
}

// FilesystemBackupManager implements BackupManager.
type FilesystemBackupManager struct {
	backupDir  string
	targetPath string
}

// NewFilesystemBackupManager creates a FilesystemBackupManager with default paths.
func NewFilesystemBackupManager() *FilesystemBackupManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &FilesystemBackupManager{
		backupDir:  filepath.Join(home, ".config", "omo-switch", "backups"),
		targetPath: filepath.Join(home, ".config", "opencode", "oh-my-openagent.json"),
	}
}

// NewFilesystemBackupManagerWithPath creates a FilesystemBackupManager with explicit paths.
func NewFilesystemBackupManagerWithPath(backupDir, targetPath string) *FilesystemBackupManager {
	return &FilesystemBackupManager{
		backupDir:  backupDir,
		targetPath: targetPath,
	}
}

// CreateBackup copies the target file to the backup directory with a timestamped name.
func (m *FilesystemBackupManager) CreateBackup() (string, error) {
	if _, err := os.Stat(m.targetPath); os.IsNotExist(err) {
		return "", fmt.Errorf("target file does not exist: %s", m.targetPath)
	}

	if err := os.MkdirAll(m.backupDir, 0o755); err != nil {
		return "", fmt.Errorf("creating backup dir: %w", err)
	}

	ts := time.Now().UTC().Format("2006-01-02T15-04-05.000Z")
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")

	fileName := fmt.Sprintf("oh-my-openagent.%s.json", ts)
	dest := filepath.Join(m.backupDir, fileName)

	if err := copyFile(m.targetPath, dest); err != nil {
		return "", fmt.Errorf("copying file: %w", err)
	}

	return dest, nil
}

// ListBackups returns all backups sorted by timestamp (newest first).
func (m *FilesystemBackupManager) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("reading backup dir: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ts, ok := parseBackupTimestamp(name)
		if !ok {
			continue
		}
		backups = append(backups, BackupInfo{
			Timestamp: ts,
			FilePath:  filepath.Join(m.backupDir, name),
			FileName:  name,
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})

	return backups, nil
}

// RestoreBackup copies a backup identified by timestamp to the target path.
func (m *FilesystemBackupManager) RestoreBackup(timestamp string) error {
	backups, err := m.ListBackups()
	if err != nil {
		return err
	}

	for _, b := range backups {
		if b.Timestamp == timestamp {
			return copyFile(b.FilePath, m.targetPath)
		}
	}

	return fmt.Errorf("backup with timestamp %q not found", timestamp)
}

func parseBackupTimestamp(fileName string) (string, bool) {
	const prefix = "oh-my-openagent."
	const suffix = ".json"

	if !strings.HasPrefix(fileName, prefix) || !strings.HasSuffix(fileName, suffix) {
		return "", false
	}

	ts := strings.TrimPrefix(fileName, prefix)
	ts = strings.TrimSuffix(ts, suffix)
	return ts, true
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

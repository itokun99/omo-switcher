package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store defines the interface for config file operations.
type Store interface {
	ListConfigs() (map[string]string, error) // alias -> filename
	GetConfig(alias string) (string, error)  // returns filename
	ReadConfig(alias string) ([]byte, error) // returns raw content
	WriteConfig(alias string, content []byte) error
	ConfigDir() string
	TargetPath() string
}

// FilesystemStore implements Store using the filesystem.
type FilesystemStore struct {
	configDir  string // ~/.config/opencode/omo_configs/
	targetPath string // ~/.config/opencode/oh-my-openagent.json
}

// NewFilesystemStore creates a FilesystemStore with default paths derived from HOME.
func NewFilesystemStore() *FilesystemStore {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &FilesystemStore{
		configDir:  filepath.Join(home, ".config", "opencode", "omo_configs"),
		targetPath: filepath.Join(home, ".config", "opencode", "oh-my-openagent.json"),
	}
}

// NewFilesystemStoreWithPath creates a FilesystemStore with explicit paths (for testing).
func NewFilesystemStoreWithPath(configDir, targetPath string) *FilesystemStore {
	return &FilesystemStore{
		configDir:  configDir,
		targetPath: targetPath,
	}
}

// ListConfigs scans configDir for omo-*.json files and returns a map of alias -> filename.
func (s *FilesystemStore) ListConfigs() (map[string]string, error) {
	entries, err := os.ReadDir(s.configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("reading config dir: %w", err)
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "omo-") && strings.HasSuffix(name, ".json") {
			alias := strings.TrimPrefix(name, "omo-")
			alias = strings.TrimSuffix(alias, ".json")
			result[alias] = name
		}
	}
	return result, nil
}

// GetConfig returns the filename for a given alias.
func (s *FilesystemStore) GetConfig(alias string) (string, error) {
	configs, err := s.ListConfigs()
	if err != nil {
		return "", err
	}
	filename, ok := configs[alias]
	if !ok {
		return "", fmt.Errorf("config %q not found", alias)
	}
	return filename, nil
}

// ReadConfig reads and returns the raw content of the config file for a given alias.
func (s *FilesystemStore) ReadConfig(alias string) ([]byte, error) {
	filename, err := s.GetConfig(alias)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(s.configDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", alias, err)
	}
	return data, nil
}

// WriteConfig writes content to the config file for a given alias.
func (s *FilesystemStore) WriteConfig(alias string, content []byte) error {
	if err := os.MkdirAll(s.configDir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	path := filepath.Join(s.configDir, fmt.Sprintf("omo-%s.json", alias))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("writing config %q: %w", alias, err)
	}
	return nil
}

// ConfigDir returns the config directory path.
func (s *FilesystemStore) ConfigDir() string {
	return s.configDir
}

// TargetPath returns the target config file path.
func (s *FilesystemStore) TargetPath() string {
	return s.targetPath
}



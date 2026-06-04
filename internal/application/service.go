// Package application contains the service layer that orchestrates
// domain models and infrastructure interfaces to implement core business logic.
package application

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/itokun99/omo-switch/internal/domain"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

// ConfigService orchestrates config operations using domain types
// and infrastructure interfaces.
type ConfigService struct {
	store     infrastructure.Store
	backup    infrastructure.BackupManager
	validator domain.SchemaValidator
}

// NewConfigService creates a ConfigService with the given dependencies.
func NewConfigService(store infrastructure.Store, backup infrastructure.BackupManager, validator domain.SchemaValidator) *ConfigService {
	return &ConfigService{
		store:     store,
		backup:    backup,
		validator: validator,
	}
}

// knownGroupNames returns the group names in display order.
func knownGroupNames() []string {
	return []string{"Mono", "Optimized", "Low-Cost", "Custom"}
}

// ListConfigs discovers all configs, validates them, and groups them
// by the known group classification. Configs that fail to read are logged
// and skipped. Returns groups in order: Mono, Optimized, Low-Cost, Custom.
func (s *ConfigService) ListConfigs() ([]domain.Group, error) {
	aliases, err := s.store.ListConfigs()
	if err != nil {
		return nil, fmt.Errorf("listing configs: %w", err)
	}

	groups := make(map[string]*domain.Group)
	for _, name := range knownGroupNames() {
		g := domain.NewGroup(name)
		groups[name] = &g
	}

	for alias, filename := range aliases {
		content, err := s.store.ReadConfig(alias)
		if err != nil {
			slog.Error("reading config", "alias", alias, "error", err)
			continue
		}

		filePath := filepath.Join(s.store.ConfigDir(), filename)
		cfg := domain.NewConfig(alias, filename, filePath, content)
		cfg = cfg.Validate(s.validator)

		groupName := domain.GetGroupForAlias(alias)
		groups[groupName].AddConfig(cfg)
	}

	result := make([]domain.Group, 0, len(groups))
	for _, name := range knownGroupNames() {
		result = append(result, *groups[name])
	}

	return result, nil
}

// GetActiveConfig compares the current target file content with each
// discovered config and returns the alias of the matching config.
// Returns an empty string if no config matches or the target doesn't exist.
func (s *ConfigService) GetActiveConfig() (string, error) {
	targetContent, err := os.ReadFile(s.store.TargetPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading target config: %w", err)
	}

	aliases, err := s.store.ListConfigs()
	if err != nil {
		return "", fmt.Errorf("listing configs: %w", err)
	}

	for alias := range aliases {
		content, err := s.store.ReadConfig(alias)
		if err != nil {
			continue
		}
		if string(content) == string(targetContent) {
			return alias, nil
		}
	}

	return "", nil
}

// SwitchConfig validates the given config, creates a backup of the current
// target, and writes the config to the target path.
func (s *ConfigService) SwitchConfig(alias string) error {
	content, err := s.store.ReadConfig(alias)
	if err != nil {
		return fmt.Errorf("reading config %q: %w", alias, err)
	}

	cfg := domain.NewConfig(alias, "", "", content)
	cfg = cfg.Validate(s.validator)
	if !cfg.IsValid {
		return fmt.Errorf("invalid config %q: %s", alias, cfg.Error)
	}

	if _, err := s.backup.CreateBackup(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	if err := os.WriteFile(s.store.TargetPath(), content, 0o644); err != nil {
		return fmt.Errorf("writing target config: %w", err)
	}

	return nil
}

// ShowConfig returns the raw content of the config with the given alias.
func (s *ConfigService) ShowConfig(alias string) (string, error) {
	content, err := s.store.ReadConfig(alias)
	if err != nil {
		return "", fmt.Errorf("reading config %q: %w", alias, err)
	}
	return string(content), nil
}

// ValidateConfig checks whether the config with the given alias passes
// schema validation. Returns (isValid, errorMessage, error).
func (s *ConfigService) ValidateConfig(alias string) (bool, string, error) {
	content, err := s.store.ReadConfig(alias)
	if err != nil {
		return false, "", fmt.Errorf("reading config %q: %w", alias, err)
	}

	cfg := domain.NewConfig(alias, "", "", content)
	cfg = cfg.Validate(s.validator)
	return cfg.IsValid, cfg.Error, nil
}

// BackupConfig creates a backup of the current target config.
// Returns the path to the backup file.
func (s *ConfigService) BackupConfig() (string, error) {
	path, err := s.backup.CreateBackup()
	if err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}
	return path, nil
}

// ListBackups returns all backups sorted by timestamp (newest first).
func (s *ConfigService) ListBackups() ([]infrastructure.BackupInfo, error) {
	return s.backup.ListBackups()
}

// RestoreBackup restores the backup identified by timestamp.
func (s *ConfigService) RestoreBackup(timestamp string) error {
	return s.backup.RestoreBackup(timestamp)
}

// GetConfigPath returns the full filesystem path for the config with the given alias.
func (s *ConfigService) GetConfigPath(alias string) (string, error) {
	filename, err := s.store.GetConfig(alias)
	if err != nil {
		return "", fmt.Errorf("getting config %q: %w", alias, err)
	}
	return filepath.Join(s.store.ConfigDir(), filename), nil
}

// ReloadConfigs forces a fresh read of all configs by delegating to ListConfigs.
func (s *ConfigService) ReloadConfigs() ([]domain.Group, error) {
	return s.ListConfigs()
}

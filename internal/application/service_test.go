package application_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/itokun99/omo-switch/internal/application"
	"github.com/itokun99/omo-switch/internal/domain"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

// mockStore implements infrastructure.Store for testing.
type mockStore struct {
	configs   map[string]string
	contents  map[string][]byte
	configDir string
	targetPath string
	listErr   error
	readErr   map[string]error
}

var _ infrastructure.Store = (*mockStore)(nil)

func (m *mockStore) ListConfigs() (map[string]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.configs, nil
}

func (m *mockStore) GetConfig(alias string) (string, error) {
	filename, ok := m.configs[alias]
	if !ok {
		return "", fmt.Errorf("config %q not found", alias)
	}
	return filename, nil
}

func (m *mockStore) ReadConfig(alias string) ([]byte, error) {
	if err, ok := m.readErr[alias]; ok {
		return nil, err
	}
	content, ok := m.contents[alias]
	if !ok {
		return nil, fmt.Errorf("config %q not found", alias)
	}
	return content, nil
}

func (m *mockStore) WriteConfig(_ string, _ []byte) error {
	return nil
}

func (m *mockStore) ConfigDir() string {
	return m.configDir
}

func (m *mockStore) TargetPath() string {
	return m.targetPath
}

// mockBackupManager implements infrastructure.BackupManager for testing.
type mockBackupManager struct {
	backupPath string
	backupErr  error
}

var _ infrastructure.BackupManager = (*mockBackupManager)(nil)

func (m *mockBackupManager) CreateBackup() (string, error) {
	if m.backupErr != nil {
		return "", m.backupErr
	}
	return m.backupPath, nil
}

func (m *mockBackupManager) ListBackups() ([]infrastructure.BackupInfo, error) {
	return nil, nil
}

func (m *mockBackupManager) RestoreBackup(_ string) error {
	return nil
}

// validJSON is a valid config with the required "agents" key.
var validJSON = []byte(`{"agents":{"sisyphus":{"model":"test"}}}`)

// invalidJSON is missing the "agents" key.
var invalidJSON = []byte(`{"categories":{"deep":{"model":"test"}}}`)

func newTestService(s *mockStore, b *mockBackupManager) *application.ConfigService {
	return application.NewConfigService(s, b, domain.DefaultValidator{})
}

func TestListConfigs(t *testing.T) {
	tests := []struct {
		name      string
		store     *mockStore
		wantCount int // expected total config count across all groups
		wantGroup string // a group that should have at least one config
	}{
		{
			name: "multiple groups",
			store: &mockStore{
				configs: map[string]string{
					"claude":         "omo-claude.json",
					"optimized-high": "omo-optimized-high.json",
					"lc-mode-low":    "omo-lc-mode-low.json",
					"my-custom":      "omo-my-custom.json",
				},
				contents: map[string][]byte{
					"claude":         validJSON,
					"optimized-high": validJSON,
					"lc-mode-low":    validJSON,
					"my-custom":      validJSON,
				},
				configDir: "/tmp/configs",
			},
			wantCount: 4,
			wantGroup: "Custom",
		},
		{
			name: "empty configs",
			store: &mockStore{
				configs:  map[string]string{},
				contents: map[string][]byte{},
				configDir: "/tmp/configs",
			},
			wantCount: 0,
		},
		{
			name: "skip read errors",
			store: &mockStore{
				configs: map[string]string{
					"claude":         "omo-claude.json",
					"optimized-high": "omo-optimized-high.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				readErr: map[string]error{
					"optimized-high": fmt.Errorf("permission denied"),
				},
				configDir: "/tmp/configs",
			},
			wantCount: 1,
		},
		{
			name: "list error",
			store: &mockStore{
				listErr:  fmt.Errorf("cannot read directory"),
				configDir: "/tmp/configs",
			},
			wantCount: -1, // signals error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.store, &mockBackupManager{})

			groups, err := svc.ListConfigs()

			if tt.wantCount == -1 {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(groups) != 4 {
				t.Fatalf("expected 4 groups, got %d", len(groups))
			}

			// Verify group order
			expectedOrder := []string{"Mono", "Optimized", "Low-Cost", "Custom"}
			for i, name := range expectedOrder {
				if groups[i].Name != name {
					t.Errorf("group[%d].Name = %q, want %q", i, groups[i].Name, name)
				}
			}

			// Count total configs
			total := 0
			for _, g := range groups {
				total += len(g.Configs)
				for _, c := range g.Configs {
					if !c.IsValid {
						t.Errorf("config %q should be valid", c.Alias)
					}
				}
			}
			if total != tt.wantCount {
				t.Errorf("total configs = %d, want %d", total, tt.wantCount)
			}

			// Verify specific group has configs
			if tt.wantGroup != "" {
				for _, g := range groups {
					if g.Name == tt.wantGroup && len(g.Configs) == 0 {
						t.Errorf("group %q should have configs", tt.wantGroup)
					}
				}
			}
		})
	}
}

func TestGetActiveConfig(t *testing.T) {
	tests := []struct {
		name          string
		setupTarget   func(dir string) // writes target file
		store         *mockStore
		wantAlias     string
	}{
		{
			name: "matching config found",
			setupTarget: func(dir string) {
				os.WriteFile(filepath.Join(dir, "target.json"), validJSON, 0o644)
			},
			store: &mockStore{
				configs: map[string]string{
					"claude":         "omo-claude.json",
					"optimized-high": "omo-optimized-high.json",
				},
				contents: map[string][]byte{
					"claude":         []byte(`{"agents":{"sisyphus":{"model":"other"}}}`),
					"optimized-high": validJSON,
				},
				configDir: "/tmp/configs",
			},
			wantAlias: "optimized-high",
		},
		{
			name: "no match",
			setupTarget: func(dir string) {
				os.WriteFile(filepath.Join(dir, "target.json"), []byte(`{"agents":{"sisyphus":{"model":"unknown"}}}`), 0o644)
			},
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			wantAlias: "",
		},
		{
			name: "target does not exist",
			setupTarget: func(dir string) {
				// no-op — target file not created
			},
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			wantAlias: "",
		},
		{
			name: "empty aliases",
			setupTarget: func(dir string) {
				os.WriteFile(filepath.Join(dir, "target.json"), validJSON, 0o644)
			},
			store: &mockStore{
				configs:  map[string]string{},
				contents: map[string][]byte{},
				configDir: "/tmp/configs",
			},
			wantAlias: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setupTarget(dir)
			tt.store.targetPath = filepath.Join(dir, "target.json")

			svc := newTestService(tt.store, &mockBackupManager{})

			got, err := svc.GetActiveConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.wantAlias {
				t.Errorf("GetActiveConfig() = %q, want %q", got, tt.wantAlias)
			}
		})
	}
}

func TestSwitchConfig(t *testing.T) {
	tests := []struct {
		name        string
		alias       string
		store       *mockStore
		backup      *mockBackupManager
		wantErr     bool
		wantContent []byte // expected content in target after switch
	}{
		{
			name:  "happy path",
			alias: "optimized-high",
			store: &mockStore{
				configs: map[string]string{
					"optimized-high": "omo-optimized-high.json",
				},
				contents: map[string][]byte{
					"optimized-high": validJSON,
				},
				configDir: "/tmp/configs",
			},
			backup: &mockBackupManager{
				backupPath: "/backups/some-file.json",
			},
			wantErr:     false,
			wantContent: validJSON,
		},
		{
			name:  "invalid config rejected",
			alias: "bad-config",
			store: &mockStore{
				configs: map[string]string{
					"bad-config": "omo-bad-config.json",
				},
				contents: map[string][]byte{
					"bad-config": invalidJSON,
				},
				configDir: "/tmp/configs",
			},
			backup:  &mockBackupManager{},
			wantErr: true,
		},
		{
			name:  "backup failure",
			alias: "optimized-high",
			store: &mockStore{
				configs: map[string]string{
					"optimized-high": "omo-optimized-high.json",
				},
				contents: map[string][]byte{
					"optimized-high": validJSON,
				},
				configDir: "/tmp/configs",
			},
			backup: &mockBackupManager{
				backupErr: fmt.Errorf("disk full"),
			},
			wantErr: true,
		},
		{
			name:  "read failure",
			alias: "missing",
			store: &mockStore{
				configs:  map[string]string{},
				contents: map[string][]byte{},
				configDir: "/tmp/configs",
			},
			backup:  &mockBackupManager{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			targetPath := filepath.Join(dir, "oh-my-openagent.json")
			tt.store.targetPath = targetPath

			svc := newTestService(tt.store, tt.backup)

			err := svc.SwitchConfig(tt.alias)

			if (err != nil) != tt.wantErr {
				t.Errorf("SwitchConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.wantContent != nil {
				got, rerr := os.ReadFile(targetPath)
				if rerr != nil {
					t.Fatalf("reading target file: %v", rerr)
				}
				if string(got) != string(tt.wantContent) {
					t.Errorf("target content = %q, want %q", got, tt.wantContent)
				}
			}
		})
	}
}

func TestShowConfig(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		store     *mockStore
		want      string
		wantErr   bool
	}{
		{
			name:  "existing config",
			alias: "claude",
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			want:    string(validJSON),
			wantErr: false,
		},
		{
			name:  "missing config",
			alias: "nonexistent",
			store: &mockStore{
				configs:  map[string]string{},
				contents: map[string][]byte{},
				configDir: "/tmp/configs",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.store, &mockBackupManager{})

			got, err := svc.ShowConfig(tt.alias)

			if (err != nil) != tt.wantErr {
				t.Errorf("ShowConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ShowConfig() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		store     *mockStore
		wantValid bool
		wantMsg   string
		wantErr   bool
	}{
		{
			name:  "valid config",
			alias: "valid-cfg",
			store: &mockStore{
				configs: map[string]string{
					"valid-cfg": "omo-valid-cfg.json",
				},
				contents: map[string][]byte{
					"valid-cfg": validJSON,
				},
				configDir: "/tmp/configs",
			},
			wantValid: true,
			wantMsg:   "",
			wantErr:   false,
		},
		{
			name:  "invalid config",
			alias: "invalid-cfg",
			store: &mockStore{
				configs: map[string]string{
					"invalid-cfg": "omo-invalid-cfg.json",
				},
				contents: map[string][]byte{
					"invalid-cfg": invalidJSON,
				},
				configDir: "/tmp/configs",
			},
			wantValid: false,
			wantMsg:   "missing required key: agents",
			wantErr:   false,
		},
		{
			name:  "missing config",
			alias: "nonexistent",
			store: &mockStore{
				configs:  map[string]string{},
				contents: map[string][]byte{},
				configDir: "/tmp/configs",
			},
			wantValid: false,
			wantMsg:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.store, &mockBackupManager{})

			valid, msg, err := svc.ValidateConfig(tt.alias)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if valid != tt.wantValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", valid, tt.wantValid)
			}
			if msg != tt.wantMsg {
				t.Errorf("ValidateConfig() msg = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestBackupConfig(t *testing.T) {
	tests := []struct {
		name     string
		backup   *mockBackupManager
		wantPath string
		wantErr  bool
	}{
		{
			name: "success",
			backup: &mockBackupManager{
				backupPath: "/backups/test.json",
			},
			wantPath: "/backups/test.json",
			wantErr:  false,
		},
		{
			name: "failure",
			backup: &mockBackupManager{
				backupErr: fmt.Errorf("disk full"),
			},
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockStore{configDir: "/tmp/configs"}, tt.backup)

			got, err := svc.BackupConfig()

			if (err != nil) != tt.wantErr {
				t.Errorf("BackupConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.wantPath {
				t.Errorf("BackupConfig() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestReloadConfigs(t *testing.T) {
	t.Run("delegates to ListConfigs", func(t *testing.T) {
		store := &mockStore{
			configs: map[string]string{
				"claude": "omo-claude.json",
			},
			contents: map[string][]byte{
				"claude": validJSON,
			},
			configDir: "/tmp/configs",
		}

		svc := newTestService(store, &mockBackupManager{})

		groups, err := svc.ReloadConfigs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(groups) != 4 {
			t.Fatalf("expected 4 groups, got %d", len(groups))
		}

		total := 0
		foundClaude := false
		for _, g := range groups {
			total += len(g.Configs)
			for _, c := range g.Configs {
				if c.Alias == "claude" {
					foundClaude = true
				}
			}
		}
		if total != 1 {
			t.Errorf("expected 1 config, got %d", total)
		}
		if !foundClaude {
			t.Error("expected to find claude config")
		}
	})
}

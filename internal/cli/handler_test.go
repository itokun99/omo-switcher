package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/itokun99/omo-switch/internal/application"
	"github.com/itokun99/omo-switch/internal/cli"
	"github.com/itokun99/omo-switch/internal/domain"
	"github.com/itokun99/omo-switch/internal/infrastructure"
)

type mockStore struct {
	configs    map[string]string
	contents   map[string][]byte
	configDir  string
	targetPath string
	listErr    error
	readErr    map[string]error
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

func (m *mockStore) WriteConfig(_ string, _ []byte) error { return nil }
func (m *mockStore) ConfigDir() string                     { return m.configDir }
func (m *mockStore) TargetPath() string                    { return m.targetPath }

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

func (m *mockBackupManager) RestoreBackup(_ string) error { return nil }

var validJSON = []byte(`{"agents":{"sisyphus":{"model":"test"}}}`)

func newTestService(store *mockStore, backup *mockBackupManager) *application.ConfigService {
	return application.NewConfigService(store, backup, domain.DefaultValidator{})
}

func setupTarget(t *testing.T, dir string, content []byte) string {
	t.Helper()
	target := filepath.Join(dir, "oh-my-openagent.json")
	if content != nil {
		if err := os.WriteFile(target, content, 0o644); err != nil {
			t.Fatalf("writing target: %v", err)
		}
	}
	return target
}

func TestHandleList(t *testing.T) {
	tests := []struct {
		name     string
		store    *mockStore
		backup   *mockBackupManager
		wantCode int
		contains []string
	}{
		{
			name: "groups displayed with active marker",
			store: &mockStore{
				configs: map[string]string{
					"claude":         "omo-claude.json",
					"optimized-high": "omo-optimized-high.json",
					"my-custom":      "omo-my-custom.json",
				},
				contents: map[string][]byte{
					"claude":         validJSON,
					"optimized-high": validJSON,
					"my-custom":      validJSON,
				},
				configDir: "/tmp/configs",
			},
			backup: &mockBackupManager{},
			wantCode: 0,
			contains: []string{
				"Available configs:",
				"Mono",
				"claude",
				"omo-claude.json",
				"Optimized",
				"optimized-high",
				"Custom",
				"my-custom",
			},
		},
		{
			name: "active config shows marker",
			store: &mockStore{
				configs: map[string]string{
					"claude":         "omo-claude.json",
					"optimized-high": "omo-optimized-high.json",
				},
				contents: map[string][]byte{
					"claude":         validJSON,
					"optimized-high": validJSON,
				},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{},
			wantCode: 0,
			contains: []string{"active"},
		},
		{
			name: "empty configs",
			store: &mockStore{
				configs:   map[string]string{},
				contents:  map[string][]byte{},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{},
			wantCode: 0,
			contains: []string{"Available configs:"},
		},
		{
			name: "list error returns exit 1",
			store: &mockStore{
				listErr:   fmt.Errorf("permission denied"),
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{},
			wantCode: 1,
			contains: []string{"Error:", "permission denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			if tt.name == "active config shows marker" {
				dir := t.TempDir()
				tt.store.targetPath = setupTarget(t, dir, validJSON)
			}

			svc := newTestService(tt.store, tt.backup)
			code := cli.Handle(svc, []string{"--list"}, &buf)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			output := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output missing %q\nGot: %s", s, output)
				}
			}
		})
	}
}

func TestHandleListNoArgs(t *testing.T) {
	store := &mockStore{
		configs:   map[string]string{"claude": "omo-claude.json"},
		contents:  map[string][]byte{"claude": validJSON},
		configDir: "/tmp/configs",
	}
	svc := newTestService(store, &mockBackupManager{})

	var buf bytes.Buffer
	code := cli.Handle(svc, nil, &buf)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "Available configs:") {
		t.Error("no-args should trigger list command")
	}
}

func TestHandleCurrent(t *testing.T) {
	tests := []struct {
		name     string
		store    *mockStore
		target   []byte
		wantCode int
		contains string
	}{
		{
			name: "active config found",
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			target:   validJSON,
			wantCode: 0,
			contains: "Current: claude",
		},
		{
			name: "no active config",
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			target:   []byte(`{"agents":{"other":"value"}}`),
			wantCode: 0,
			contains: "No active config matched",
		},
		{
			name: "target missing",
			store: &mockStore{
				configs: map[string]string{
					"claude": "omo-claude.json",
				},
				contents: map[string][]byte{
					"claude": validJSON,
				},
				configDir: "/tmp/configs",
			},
			target:   nil,
			wantCode: 0,
			contains: "No active config matched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.store.targetPath = setupTarget(t, dir, tt.target)

			svc := newTestService(tt.store, &mockBackupManager{})
			var buf bytes.Buffer

			code := cli.Handle(svc, []string{"--current"}, &buf)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("output = %q, want containing %q", buf.String(), tt.contains)
			}
		})
	}
}

func TestHandleCurrentShortFlag(t *testing.T) {
	dir := t.TempDir()
	store := &mockStore{
		configs:   map[string]string{"claude": "omo-claude.json"},
		contents:  map[string][]byte{"claude": validJSON},
		configDir: "/tmp/configs",
		targetPath: setupTarget(t, dir, validJSON),
	}
	svc := newTestService(store, &mockBackupManager{})

	var buf bytes.Buffer
	code := cli.Handle(svc, []string{"-c"}, &buf)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "Current:") {
		t.Error("-c should trigger current command")
	}
}

func TestHandleHelp(t *testing.T) {
	svc := newTestService(&mockStore{configDir: "/tmp"}, &mockBackupManager{})
	var buf bytes.Buffer

	code := cli.Handle(svc, []string{"--help"}, &buf)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	output := buf.String()
	required := []string{
		"Usage: omo-switch",
		"--list, -l",
		"--current, -c",
		"show <alias>",
		"Examples:",
	}
	for _, s := range required {
		if !strings.Contains(output, s) {
			t.Errorf("help output missing %q", s)
		}
	}
}

func TestHandleHelpShortFlag(t *testing.T) {
	svc := newTestService(&mockStore{configDir: "/tmp"}, &mockBackupManager{})
	var buf bytes.Buffer

	code := cli.Handle(svc, []string{"-h"}, &buf)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Error("-h should trigger help")
	}
}

func TestHandleShow(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		store    *mockStore
		wantCode int
		contains []string
	}{
		{
			name:  "existing config",
			alias: "claude",
			store: &mockStore{
				configs:   map[string]string{"claude": "omo-claude.json"},
				contents:  map[string][]byte{"claude": validJSON},
				configDir: "/tmp/configs",
			},
			wantCode: 0,
			contains: []string{"--- claude (omo-claude.json) ---", string(validJSON)},
		},
		{
			name:  "missing config",
			alias: "nonexistent",
			store: &mockStore{
				configs:   map[string]string{},
				contents:  map[string][]byte{},
				configDir: "/tmp/configs",
			},
			wantCode: 1,
			contains: []string{"Error:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(tt.store, &mockBackupManager{})
			var buf bytes.Buffer

			code := cli.Handle(svc, []string{"show", tt.alias}, &buf)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			output := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output missing %q\nGot: %s", s, output)
				}
			}
		})
	}
}

func TestHandleShowMissingArg(t *testing.T) {
	svc := newTestService(&mockStore{configDir: "/tmp"}, &mockBackupManager{})
	var buf bytes.Buffer

	code := cli.Handle(svc, []string{"show"}, &buf)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(buf.String(), "Error:") {
		t.Error("expected error message for missing alias")
	}
}

func TestHandleSwitch(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		store    *mockStore
		backup   *mockBackupManager
		wantCode int
		contains string
	}{
		{
			name:  "successful switch",
			alias: "claude",
			store: &mockStore{
				configs:   map[string]string{"claude": "omo-claude.json"},
				contents:  map[string][]byte{"claude": validJSON},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{backupPath: "/backups/test.json"},
			wantCode: 0,
			contains: "Switched to: claude",
		},
		{
			name:  "invalid config",
			alias: "bad",
			store: &mockStore{
				configs:   map[string]string{"bad": "omo-bad.json"},
				contents:  map[string][]byte{"bad": []byte(`{"no_agents": true}`)},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{},
			wantCode: 1,
			contains: "Error:",
		},
		{
			name:  "missing config",
			alias: "nonexistent",
			store: &mockStore{
				configs:   map[string]string{},
				contents:  map[string][]byte{},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{},
			wantCode: 1,
			contains: "Error:",
		},
		{
			name:  "backup failure",
			alias: "claude",
			store: &mockStore{
				configs:   map[string]string{"claude": "omo-claude.json"},
				contents:  map[string][]byte{"claude": validJSON},
				configDir: "/tmp/configs",
			},
			backup:   &mockBackupManager{backupErr: fmt.Errorf("disk full")},
			wantCode: 1,
			contains: "Error:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.store.targetPath = filepath.Join(dir, "oh-my-openagent.json")

			svc := newTestService(tt.store, tt.backup)
			var buf bytes.Buffer

			code := cli.Handle(svc, []string{tt.alias}, &buf)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("output = %q, want containing %q", buf.String(), tt.contains)
			}
		})
	}
}

func TestHandleUnknownFlag(t *testing.T) {
	svc := newTestService(&mockStore{configDir: "/tmp"}, &mockBackupManager{})
	var buf bytes.Buffer

	code := cli.Handle(svc, []string{"--bogus"}, &buf)

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	output := buf.String()
	if !strings.Contains(output, "unknown option") {
		t.Errorf("output = %q, want containing 'unknown option'", output)
	}
	if !strings.Contains(output, "--help") {
		t.Error("should suggest running --help")
	}
}

func TestHandleListSortedOrder(t *testing.T) {
	store := &mockStore{
		configs: map[string]string{
			"claude":         "omo-claude.json",
			"minimax":        "omo-minimax.json",
			"optimized-high": "omo-optimized-high.json",
		},
		contents: map[string][]byte{
			"claude":         validJSON,
			"minimax":        validJSON,
			"optimized-high": validJSON,
		},
		configDir: "/tmp/configs",
	}
	svc := newTestService(store, &mockBackupManager{})
	var buf bytes.Buffer

	cli.Handle(svc, []string{"--list"}, &buf)
	output := buf.String()

	claudeIdx := strings.Index(output, "claude")
	minimaxIdx := strings.Index(output, "minimax")

	if claudeIdx < 0 || minimaxIdx < 0 {
		t.Fatalf("expected both claude and minimax in output:\n%s", output)
	}
	if minimaxIdx >= claudeIdx {
		t.Errorf("minimax should appear before claude in Mono group (known order)")
	}
}

func TestHandleListCustomSorted(t *testing.T) {
	store := &mockStore{
		configs: map[string]string{
			"zebra":  "omo-zebra.json",
			"alpha":  "omo-alpha.json",
			"middle": "omo-middle.json",
		},
		contents: map[string][]byte{
			"zebra":  validJSON,
			"alpha":  validJSON,
			"middle": validJSON,
		},
		configDir: "/tmp/configs",
	}
	svc := newTestService(store, &mockBackupManager{})
	var buf bytes.Buffer

	cli.Handle(svc, []string{"--list"}, &buf)
	output := buf.String()

	alphaIdx := strings.Index(output, "alpha")
	middleIdx := strings.Index(output, "middle")
	zebraIdx := strings.Index(output, "zebra")

	if alphaIdx < 0 || middleIdx < 0 || zebraIdx < 0 {
		t.Fatalf("expected all custom configs in output:\n%s", output)
	}
	if !(alphaIdx < middleIdx && middleIdx < zebraIdx) {
		t.Errorf("custom configs should be sorted alphabetically")
	}
}

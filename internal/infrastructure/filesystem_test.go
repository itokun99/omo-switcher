package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemStore_ListConfigs(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "discovers omo-*.json files",
			files: []string{"omo-high.json", "omo-low.json", "other.txt"},
			want:  map[string]string{"high": "omo-high.json", "low": "omo-low.json"},
		},
		{
			name:  "ignores non-matching files",
			files: []string{"config.json", "omo-test.txt", "readme.md"},
			want:  map[string]string{},
		},
		{
			name:  "empty directory",
			files: []string{},
			want:  map[string]string{},
		},
		{
			name:  "nonexistent directory returns empty map",
			files: nil,
			want:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.files != nil {
				for _, f := range tt.files {
					writeTestFile(t, filepath.Join(dir, f), "{}")
				}
			} else {
				dir = filepath.Join(dir, "nonexistent")
			}

			store := NewFilesystemStoreWithPath(dir, filepath.Join(dir, "target.json"))
			got, err := store.ListConfigs()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !mapsEqual(got, tt.want) {
				t.Errorf("ListConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilesystemStore_GetConfig(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "omo-high.json"), "{}")

	store := NewFilesystemStoreWithPath(dir, filepath.Join(dir, "target.json"))

	t.Run("existing config", func(t *testing.T) {
		got, err := store.GetConfig("high")
		if err != nil {
			t.Fatalf("GetConfig() error = %v", err)
		}
		if got != "omo-high.json" {
			t.Errorf("GetConfig() = %q, want %q", got, "omo-high.json")
		}
	})

	t.Run("missing config", func(t *testing.T) {
		_, err := store.GetConfig("nonexistent")
		if err == nil {
			t.Fatal("GetConfig() expected error for missing config")
		}
	})
}

func TestFilesystemStore_ReadConfig(t *testing.T) {
	dir := t.TempDir()
	content := `{"agents":{"sisyphus":{"model":"test"}}}`
	writeTestFile(t, filepath.Join(dir, "omo-test.json"), content)

	store := NewFilesystemStoreWithPath(dir, filepath.Join(dir, "target.json"))

	t.Run("reads file content", func(t *testing.T) {
		got, err := store.ReadConfig("test")
		if err != nil {
			t.Fatalf("ReadConfig() error = %v", err)
		}
		if string(got) != content {
			t.Errorf("ReadConfig() = %q, want %q", string(got), content)
		}
	})

	t.Run("missing config returns error", func(t *testing.T) {
		_, err := store.ReadConfig("nonexistent")
		if err == nil {
			t.Fatal("ReadConfig() expected error for missing config")
		}
	})
}

func TestFilesystemStore_WriteConfig(t *testing.T) {
	dir := t.TempDir()
	store := NewFilesystemStoreWithPath(dir, filepath.Join(dir, "target.json"))

	content := []byte(`{"new": "config"}`)
	if err := store.WriteConfig("created", content); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "omo-created.json"))
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("written content = %q, want %q", string(got), string(content))
	}
}

func TestFilesystemStore_Getters(t *testing.T) {
	store := NewFilesystemStoreWithPath("/tmp/configs", "/tmp/target.json")
	if store.ConfigDir() != "/tmp/configs" {
		t.Errorf("ConfigDir() = %q, want %q", store.ConfigDir(), "/tmp/configs")
	}
	if store.TargetPath() != "/tmp/target.json" {
		t.Errorf("TargetPath() = %q, want %q", store.TargetPath(), "/tmp/target.json")
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

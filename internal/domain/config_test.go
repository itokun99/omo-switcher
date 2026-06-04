package domain

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		fileName string
		filePath string
		content  []byte
	}{
		{
			name:     "basic config",
			alias:    "claude",
			fileName: "omo-claude.json",
			filePath: "/home/user/.config/opencode/omo_configs/omo-claude.json",
			content:  []byte(`{"agents":{"sisyphus":{"model":"test"}}}`),
		},
		{
			name:     "empty content",
			alias:    "empty",
			fileName: "omo-empty.json",
			filePath: "/tmp/omo-empty.json",
			content:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.alias, tt.fileName, tt.filePath, tt.content)

			if got.Alias != tt.alias {
				t.Errorf("Alias = %q, want %q", got.Alias, tt.alias)
			}
			if got.FileName != tt.fileName {
				t.Errorf("FileName = %q, want %q", got.FileName, tt.fileName)
			}
			if got.FilePath != tt.filePath {
				t.Errorf("FilePath = %q, want %q", got.FilePath, tt.filePath)
			}
			if string(got.Content) != string(tt.content) {
				t.Errorf("Content = %q, want %q", got.Content, tt.content)
			}
			if got.IsValid {
				t.Error("IsValid should default to false")
			}
			if got.Error != "" {
				t.Errorf("Error = %q, want empty", got.Error)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	validator := DefaultValidator{}

	tests := []struct {
		name      string
		content   []byte
		wantValid bool
		wantErr   string
	}{
		{
			name:      "valid config with agents key",
			content:   []byte(`{"agents":{"sisyphus":{"model":"test"}}}`),
			wantValid: true,
			wantErr:   "",
		},
		{
			name:      "invalid json",
			content:   []byte(`{invalid`),
			wantValid: false,
			wantErr:   "invalid json",
		},
		{
			name:      "missing agents key",
			content:   []byte(`{"categories":{"deep":{"model":"test"}}}`),
			wantValid: false,
			wantErr:   "missing required key: agents",
		},
		{
			name:      "empty json object",
			content:   []byte(`{}`),
			wantValid: false,
			wantErr:   "missing required key: agents",
		},
		{
			name:      "null content",
			content:   nil,
			wantValid: false,
			wantErr:   "invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewConfig("test", "test.json", "/tmp/test.json", tt.content)
			got := original.Validate(validator)

			if got.IsValid != tt.wantValid {
				t.Errorf("IsValid = %v, want %v", got.IsValid, tt.wantValid)
			}

			if tt.wantErr == "" {
				if got.Error != "" {
					t.Errorf("Error = %q, want empty", got.Error)
				}
			} else {
				if got.Error == "" {
					t.Error("Error should not be empty")
				}
			}

			// Verify immutability — original unchanged
			if original.IsValid {
				t.Error("original.IsValid should remain false")
			}
			if original.Error != "" {
				t.Error("original.Error should remain empty")
			}
		})
	}
}

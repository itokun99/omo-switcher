package domain

import (
	"testing"
)

func TestDefaultValidator_Validate(t *testing.T) {
	v := DefaultValidator{}

	tests := []struct {
		name    string
		content []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid json with agents key",
			content: []byte(`{"agents":{"sisyphus":{"model":"test"}}}`),
			wantErr: false,
		},
		{
			name:    "valid json with agents and other keys",
			content: []byte(`{"agents":{"sisyphus":{"model":"test"}},"categories":{"deep":{"model":"test"}}}`),
			wantErr: false,
		},
		{
			name:    "invalid json syntax",
			content: []byte(`{invalid json`),
			wantErr: true,
			errMsg:  "invalid json",
		},
		{
			name:    "valid json but missing agents key",
			content: []byte(`{"categories":{"deep":{"model":"test"}}}`),
			wantErr: true,
			errMsg:  "missing required key: agents",
		},
		{
			name:    "empty json object",
			content: []byte(`{}`),
			wantErr: true,
			errMsg:  "missing required key: agents",
		},
		{
			name:    "json array instead of object",
			content: []byte(`[1, 2, 3]`),
			wantErr: true,
			errMsg:  "invalid json",
		},
		{
			name:    "json string instead of object",
			content: []byte(`"hello"`),
			wantErr: true,
			errMsg:  "invalid json",
		},
		{
			name:    "json number instead of object",
			content: []byte(`42`),
			wantErr: true,
			errMsg:  "invalid json",
		},
		{
			name:    "null content",
			content: nil,
			wantErr: true,
			errMsg:  "invalid json",
		},
		{
			name:    "empty content",
			content: []byte(``),
			wantErr: true,
			errMsg:  "invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" {
					got := err.Error()
					if len(got) < len(tt.errMsg) || got[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("error = %q, want prefix %q", got, tt.errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDefaultValidator_RequiredKeys(t *testing.T) {
	v := DefaultValidator{}
	keys := v.RequiredKeys()

	if len(keys) != 1 {
		t.Fatalf("RequiredKeys() returned %d keys, want 1", len(keys))
	}
	if keys[0] != "agents" {
		t.Errorf("RequiredKeys()[0] = %q, want %q", keys[0], "agents")
	}
}

func TestSchemaValidator_Interface(t *testing.T) {
	var _ SchemaValidator = DefaultValidator{}
}

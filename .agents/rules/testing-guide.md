# Testing Guide - omo-switch

## Overview

This document defines testing standards for the omo-switch project.

## Test Framework

**Standard library `testing` only** - no testify, gomega, or external frameworks.

```go
import "testing"
```

## Test File Organization

| Layer | File | Package | Access |
|-------|------|---------|--------|
| Domain | `*_test.go` | `domain` | Unexported |
| Infrastructure | `*_test.go` | `infrastructure` | Unexported |
| Application | `*_test.go` | `application_test` | Exported only |
| CLI | `*_test.go` | `cli_test` | Exported only |
| TUI | `*_test.go` | `tui` | Unexported |
| Components | `*_test.go` | `components` | Unexported |

## Test Naming Conventions

```go
func TestNewConfig(t *testing.T)                    // Constructor
func TestConfig_Validate(t *testing.T)              // Method
func TestConfig_Validate_InvalidJSON(t *testing.T)  // Specific case
func TestFilesystemStore_ListConfigs(t *testing.T)  // Interface method
```

## Table-Driven Tests

### Pattern

```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Foo(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Foo() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Foo() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Mock Pattern

### Hand-Written Mocks

```go
type mockStore struct {
    configs map[string]string
    content map[string][]byte
    listErr error
    readErr map[string]error
}

func (m *mockStore) ListConfigs() (map[string]string, error) {
    if m.listErr != nil {
        return nil, m.listErr
    }
    return m.configs, nil
}

func (m *mockStore) ReadConfig(alias string) ([]byte, error) {
    if err, ok := m.readErr[alias]; ok {
        return nil, err
    }
    return m.content[alias], nil
}

// Compile-time interface check
var _ infrastructure.Store = (*mockStore)(nil)
```

### Error Injection

```go
type mockStore struct {
    listErr error      // Error for ListConfigs
    readErr map[string]error  // Per-config errors
    // ...
}
```

## Test Helpers

### Helper Function Pattern

```go
func writeTestFile(t *testing.T, path, content string) {
    t.Helper()
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        t.Fatal(err)
    }
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        t.Fatal(err)
    }
}
```

### Service Constructor for Tests

```go
func newTestService(store *mockStore, backup *mockBackupManager) *application.ConfigService {
    return application.NewConfigService(store, backup, domain.DefaultValidator{})
}
```

## Fixtures

### Inline Fixtures

```go
var (
    validJSON   = []byte(`{"agents": {"sisyphus": {}}}`)
    invalidJSON = []byte(`{"invalid": true}`)
)
```

### Temporary Directories

```go
func TestFilesystemStore_ListConfigs(t *testing.T) {
    dir := t.TempDir()
    writeTestFile(t, filepath.Join(dir, "omo-test.json"), `{"agents":{}}`)

    store := infrastructure.NewFilesystemStoreWithPath(dir, "")
    configs, err := store.ListConfigs()
    // ...
}
```

## Coverage

### Running Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage Expectations

- Domain layer: 100% coverage expected
- Infrastructure layer: 90%+ coverage expected
- Application layer: 90%+ coverage expected
- CLI layer: 80%+ coverage expected
- TUI layer: 70%+ coverage expected

## Testing Each Layer

### Domain Tests

```go
func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name      string
        content   []byte
        wantValid bool
    }{
        {name: "valid", content: validJSON, wantValid: true},
        {name: "invalid json", content: []byte("bad"), wantValid: false},
        {name: "missing agents", content: []byte("{}"), wantValid: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := domain.NewConfig("test", "test.json", "/path", tt.content)
            cfg = cfg.Validate(domain.DefaultValidator{})
            if cfg.IsValid != tt.wantValid {
                t.Errorf("IsValid = %v, want %v", cfg.IsValid, tt.wantValid)
            }
        })
    }
}
```

### Infrastructure Tests

```go
func TestFilesystemStore_ListConfigs(t *testing.T) {
    tests := []struct {
        name    string
        files   []string
        want    map[string]string
        wantErr bool
    }{
        {
            name:  "discovers omo-*.json files",
            files: []string{"omo-test.json", "other.json"},
            want:  map[string]string{"test": "omo-test.json"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := t.TempDir()
            for _, f := range tt.files {
                writeTestFile(t, filepath.Join(dir, f), "{}")
            }

            store := infrastructure.NewFilesystemStoreWithPath(dir, "")
            got, err := store.ListConfigs()
            // ...
        })
    }
}
```

### Application Tests

```go
func TestListConfigs(t *testing.T) {
    mock := &mockStore{
        configs: map[string]string{"test": "omo-test.json"},
        content: map[string][]byte{"test": validJSON},
    }

    service := newTestService(mock, nil)
    groups, err := service.ListConfigs()
    // ...
}
```

### CLI Tests

```go
func TestHandleList(t *testing.T) {
    mock := &mockStore{
        configs: map[string]string{"test": "omo-test.json"},
        content: map[string][]byte{"test": validJSON},
    }

    service := newTestService(mock, nil)
    var buf bytes.Buffer
    code := cli.Handle(service, []string{"--list"}, &buf)

    if code != 0 {
        t.Errorf("Handle() = %d, want 0", code)
    }
    // ...
}
```

### TUI Tests

```go
func TestApp_Update(t *testing.T) {
    mock := &mockConfigService{
        groups: []domain.Group{...},
        active: "test",
    }

    app := tui.NewApp(mock)
    // Test Update() with various messages
}
```

## Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/domain/...

# Verbose
go test -v ./...

# Specific test
go test -run TestConfig_Validate ./internal/domain/...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...
```

## Common Patterns

### Testing Error Cases

```go
t.Run("error case", func(t *testing.T) {
    mock := &mockStore{listErr: errors.New("error")}
    service := newTestService(mock, nil)

    _, err := service.ListConfigs()
    if err == nil {
        t.Error("expected error, got nil")
    }
})
```

### Testing with Temp Directories

```go
func TestFoo(t *testing.T) {
    dir := t.TempDir()
    // Use dir for test files
    // Auto-cleaned after test
}
```

### Testing JSON Content

```go
var validJSON = []byte(`{
    "agents": {
        "sisyphus": {
            "model": "opencode-go/kimi-k2.6"
        }
    }
}`)
```

# Go Engineering Standards - omo-switch

## Overview

This document defines Go coding standards specific to the omo-switch project.

## Package Design

### Package Responsibilities

| Package | Responsibility | I/O | Dependencies |
|---------|---------------|-----|--------------|
| domain | Pure business logic | None | None |
| application | Orchestration | Via interfaces | domain, infrastructure |
| infrastructure | I/O implementations | Filesystem | domain (types only) |
| cli | CLI handler | stdout | application, domain |
| tui | TUI interface | stdin/stdout | application, domain, infrastructure |

### Package Boundaries

```
domain/ → NO imports from other internal packages
application/ → imports domain/ + infrastructure/
infrastructure/ → imports domain/ types only
cli/ → imports application/ + domain/
tui/ → imports application/ + domain/ + infrastructure/
```

### Package Naming

- Use lowercase, single-word names
- Avoid `utils`, `helpers`, `common`
- Package name matches directory name

## Error Handling

### Error Creation

```go
// Always wrap with context
fmt.Errorf("listing configs: %w", err)

// Pattern: "verb-ing noun: %w"
fmt.Errorf("reading config %q: %w", alias, err)
fmt.Errorf("writing target config: %w", err)
fmt.Errorf("backup failed: %w", err)
```

### Error Propagation

```
Infrastructure → Application → CLI/TUI
  ↓                ↓              ↓
wrap system    add context    user-facing
errors         with %w        message
```

### User-Facing Errors

```go
// CLI
fmt.Fprintf(w, "Error: %v\n", err)
return 1

// TUI
a.status.SetMessage("Error: " + msg.err.Error())
```

### Graceful Degradation

```go
// Use os.IsNotExist for optional resources
if os.IsNotExist(err) {
    return emptyResult, nil
}
```

## Logging

### Logging Library

Use `log/slog` (standard library):

```go
slog.Error("reading config", "alias", alias, "error", err)
```

### Logging Levels

- `slog.Error` - Errors that need attention
- `slog.Warn` - Potential issues
- `slog.Info` - Important operations
- `slog.Debug` - Detailed debugging

### Logging Context

```go
slog.Error("operation failed",
    "operation", "list_configs",
    "error", err,
)
```

## Configuration

### Config Paths

```go
// Default paths
home, _ := os.UserHomeDir()
configDir := filepath.Join(home, ".config", "opencode", "omo_configs")
targetPath := filepath.Join(home, ".config", "opencode", "oh-my-openagent.json")
backupDir := filepath.Join(home, ".config", "omo-switch", "backups")
```

### Config File Naming

```go
// Pattern: omo-<alias>.json
filename := fmt.Sprintf("omo-%s.json", alias)
```

## Struct Design

### Constructor Pattern

```go
type Foo struct {
    field1 string
    field2 int
}

func NewFoo(field1 string, field2 int) *Foo {
    return &Foo{
        field1: field1,
        field2: field2,
    }
}
```

### Immutable Validation

```go
// Validate returns a new copy, doesn't modify original
func (c Config) Validate(validator SchemaValidator) Config {
    // Create new Config with validation result
    return Config{
        Alias:    c.Alias,
        FileName: c.FileName,
        // ...
        IsValid:  true,
    }
}
```

## Interface Design

### Interface Location

Define interfaces in the package that USES them:

```go
// infrastructure/filesystem.go - defines Store interface
type Store interface {
    ListConfigs() (map[string]string, error)
    // ...
}

// application/service.go - uses Store interface
type ConfigService struct {
    store infrastructure.Store
    // ...
}
```

### Compile-Time Checks

```go
var _ SchemaValidator = DefaultValidator{}
var _ Store = (*FilesystemStore)(nil)
```

## Naming Conventions

### Variables

- Use camelCase for local variables
- Use descriptive names (not single letters except loops)
- Avoid abbreviations unless well-known

### Functions

- Use CamelCase for exported functions
- Use camelCase for unexported functions
- Prefix command functions with `cmd` (CLI)

### Constants

- Use CamelCase for exported constants
- Use camelCase for unexported constants

### Packages

- Use lowercase, single-word names
- Avoid underscores in package names

## Code Organization

### File Structure

```go
package foo

// 1. Imports
import (
    "standard"
    "third-party"
    "internal"
)

// 2. Constants
const (
    // ...
)

// 3. Types
type Foo struct {
    // ...
}

// 4. Constructor
func NewFoo() *Foo {
    // ...
}

// 5. Methods
func (f *Foo) Bar() {
    // ...
}
```

### Import Grouping

```go
import (
    // Standard library
    "fmt"
    "os"

    // Third-party
    "github.com/charmbracelet/bubbletea"

    // Internal
    "github.com/itokun99/omo-switch/internal/domain"
)
```

## Testing

### Test File Naming

- Same directory as source: `foo_test.go`
- Same package for unexported access
- External package (`_test` suffix) for black-box testing

### Test Function Naming

```go
func TestFoo(t *testing.T)           // Simple
func TestFoo_Bar(t *testing.T)       // Method
func TestFoo_Bar_Scenario(t *testing.T) // Specific case
```

### Table-Driven Tests

```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "valid", input: "a", want: "b", wantErr: false},
        {name: "invalid", input: "", want: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Foo(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Foo() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Foo() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Build and Run

### Build

```bash
go build -o omo-switch ./cmd/omo-switch
```

### Run Tests

```bash
go test ./...
go test -cover ./...
go test -v ./internal/domain/...
```

### Lint

```bash
go vet ./...
```

## Dependencies

### Adding Dependencies

```bash
go get github.com/package/name
go mod tidy
```

### Dependency Rules

- Prefer standard library when possible
- Use well-maintained packages
- Avoid dependencies with many transitive deps
- Document why each dependency is needed

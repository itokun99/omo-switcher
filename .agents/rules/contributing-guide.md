# Contributing Guide - omo-switch

## Welcome

This guide helps new developers (human or AI) start contributing to omo-switch.

## Prerequisites

- Go 1.26.3 or later
- Git
- Terminal (for TUI testing)

## Getting Started

### 1. Clone Repository

```bash
git clone https://github.com/itokun99/omo-switch.git
cd omo-switch
```

### 2. Build

```bash
go build -o omo-switch ./cmd/omo-switch
```

### 3. Run Tests

```bash
go test ./...
```

### 4. Run Application

```bash
# TUI mode (default)
./omo-switch

# CLI mode
./omo-switch --list
./omo-switch --current
./omo-switch show claude
./omo-switch claude
```

## Project Structure

```
omo-switch/
├── cmd/omo-switch/          # Entry point
│   └── main.go
├── internal/
│   ├── domain/              # Business logic (no I/O)
│   ├── application/         # Service orchestration
│   ├── infrastructure/      # I/O implementations
│   ├── cli/                 # CLI handler
│   └── tui/                 # TUI interface
│       └── components/      # UI components
├── .agents/rules/           # AI agent documentation
└── AGENTS.md                # Master AI rules
```

## Development Workflow

### 1. Pick a Task

- Check issues for bugs/features
- Or identify improvement area

### 2. Create Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-fix
```

### 3. Implement

Follow the patterns:
- Read AGENTS.md for architecture rules
- Read relevant documentation in .agents/rules/
- Find similar existing code
- Copy patterns, don't invent new ones

### 4. Test

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/domain/...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

### 5. Verify

```bash
# Lint
go vet ./...

# Build
go build -o omo-switch ./cmd/omo-switch

# Manual test
./omo-switch --list
```

### 6. Commit

```bash
git add .
git commit -m "feat: add new feature"
# or
git commit -m "fix: fix bug description"
```

### 7. Push and PR

```bash
git push origin feature/my-feature
```

## Code Style

### Go Standards

- Follow Go conventions (gofmt, go vet)
- Use table-driven tests
- Follow existing patterns
- Don't add dependencies without strong reason

### Naming

- Package names: lowercase, single word
- Function names: CamelCase (exported), camelCase (unexported)
- Variable names: camelCase
- Constants: CamelCase (exported), camelCase (unexported)

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}

// User-facing errors
fmt.Fprintf(w, "Error: %v\n", err)
return 1
```

### Testing

```go
// Table-driven tests
func TestFoo(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "case 1", input: "a", want: "b"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Foo(tt.input)
            if got != tt.want {
                t.Errorf("Foo() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Adding Features

### New CLI Command

1. Read command-guidelines.md
2. Add case in handler.go
3. Create cmdXxx() function
4. Update help text
5. Add tests

### New TUI Component

1. Read cli-architecture.md (TUI section)
2. Create component in components/
3. Follow FooModel/FooStyles pattern
4. Add to app.go
5. Create tests

### New Config Group

1. Edit domain/group.go
2. Update KnownGroups map
3. Update tests

## Common Issues

### Tests Failing

```bash
# Run specific test
go test -run TestName ./internal/package/...

# Verbose output
go test -v ./internal/package/...

# Check for race conditions
go test -race ./...
```

### Build Errors

```bash
# Clean and rebuild
go clean
go build ./cmd/omo-switch
```

### Import Errors

```bash
# Tidy modules
go mod tidy
```

## Architecture Rules

### DO

- Follow clean architecture boundaries
- Use interfaces for dependencies
- Write table-driven tests
- Handle errors with context
- Use existing patterns

### DON'T

- Add I/O to domain layer
- Add business logic to infrastructure
- Use external test frameworks
- Share styles across TUI components
- Create new packages without strong reason
- Use Cobra or other CLI frameworks

## Getting Help

1. Read AGENTS.md
2. Read relevant .agents/rules/ documentation
3. Look at similar existing code
4. Check test files for examples

## Code Review Checklist

Before submitting PR:

- [ ] All tests pass (`go test ./...`)
- [ ] No vet warnings (`go vet ./...`)
- [ ] Build succeeds (`go build ./cmd/omo-switch`)
- [ ] Package boundaries respected
- [ ] Existing patterns followed
- [ ] Tests added for new code
- [ ] Error handling with context
- [ ] No unnecessary dependencies

## Release Process

1. Update version (if applicable)
2. Update CHANGELOG.md
3. Create git tag
4. GitHub Actions builds binaries
5. Create GitHub release

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [AGENTS.md](../AGENTS.md) - Master AI rules

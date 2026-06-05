# AGENTS.md - AI Agent Operating Rules for omo-switch

## Project Identity

**Name**: omo-switch
**Type**: CLI/TUI config switcher for oh-my-openagent
**Language**: Go 1.26.3
**Framework**: Charm ecosystem (Bubble Tea + Bubbles + Lipgloss)
**Architecture**: Clean Architecture (domain → application → infrastructure → cli/tui)

## Core Rules for AI Agents

### Rule 1: Respect Package Boundaries

```
domain/ → NO I/O, NO imports from other internal packages
application/ → imports domain/ + infrastructure/
infrastructure/ → imports domain/ types only for validation
cli/ → imports application/ + domain/
tui/ → imports application/ + domain/ + infrastructure/ (for BackupInfo only)
```

**VIOLATION EVIDENCE**: If you add `os.ReadFile` to `domain/`, you broke the architecture.

### Rule 2: Follow Interface-Based DI

All infrastructure dependencies are injected via interfaces:
- `infrastructure.Store` - config file operations
- `infrastructure.BackupManager` - backup operations
- `domain.SchemaValidator` - validation logic

**NEVER** create concrete dependencies inside service layer.

### Rule 3: CLI Command Pattern

Commands are defined in `internal/cli/handler.go` using a switch statement:
```go
func Handle(service *application.ConfigService, args []string, w io.Writer) int {
    switch args[0] {
    case "--list", "-l":
        return cmdList(service, w)
    // ...
    }
}
```

**NO Cobra, NO urfave/cli** - this project uses manual dispatch.

### Rule 4: TUI Component Pattern

Every component in `internal/tui/components/` follows this pattern:
```go
type FooModel struct { /* state */ }
type FooStyles struct { /* lipgloss styles */ }
func NewFooModel(...) FooModel
func (m *FooModel) Render(styles FooStyles, width int) string
func (m *FooModel) Show(...)
func (m *FooModel) Hide()
func (m FooModel) IsActive() bool
```

**CRITICAL**: Each component owns its own `Styles` struct. Never share styles across components.

### Rule 5: Testing Conventions

- **Framework**: Standard `testing` package only (no testify)
- **Pattern**: Table-driven with `t.Run()`
- **Mocks**: Hand-written structs with compile-time interface checks
- **Package strategy**: Same-package for unexported access, external-test for black-box

### Rule 6: Error Handling

```go
// Creation
fmt.Errorf("context: %w", err)

// User-facing (CLI)
fmt.Fprintf(w, "Error: %v\n", err)
return 1

// User-facing (TUI)
a.status.SetMessage("Error: " + msg.err.Error())
```

**NO custom error types, NO sentinel errors** in current codebase.

### Rule 7: Config File Conventions

- Config files: `~/.config/opencode/omo_configs/omo-*.json`
- Active config: `~/.config/opencode/oh-my-openagent.json`
- Backups: `~/.config/omo-switch/backups/oh-my-openagent.<timestamp>.json`
- Naming: `omo-<alias>.json` where alias is the config name

### Rule 8: Group Classification

Config groups are defined in `internal/domain/group.go`:
```go
var KnownGroups = map[string][]string{
    "Mono":      {"minimax", "qwen", "deepseek", "glm", "gpt", "claude"},
    "Optimized": {"optimized-high", "optimized-medium", "optimized-low"},
    "Low-Cost":  {"lc-mode-low", "lc-mode-medium", "lc-mode-high", "lc-mode-ultra"},
}
```

Unknown aliases fall into "Custom" group.

## Architecture Violations to Avoid

| Violation | Why It's Wrong | Where It Happens |
|-----------|---------------|------------------|
| Adding I/O to domain/ | Breaks pure business logic isolation | domain/config.go, domain/group.go |
| Adding business logic to infrastructure/ | Violates single responsibility | infrastructure/filesystem.go |
| Using external test frameworks | Inconsistent with project conventions | Any *_test.go |
| Sharing Styles structs across components | Causes circular imports | tui/components/* |
| Creating new packages without strong reason | Over-engineering for small project | internal/* |
| Using Cobra or other CLI frameworks | Project uses manual dispatch | internal/cli/handler.go |

## Technical Debt Inventory

1. **Duplicate path constants**: `targetPath` hardcoded in both `FilesystemStore` and `FilesystemBackupManager`
2. **No XDG compliance**: Hardcoded `~/.config/` paths instead of `$XDG_CONFIG_HOME`
3. **No shared test utilities**: Mocks duplicated between `application_test` and `cli_test`
4. **No cmd/ tests**: Entry point is untested
5. **Simplistic diff algorithm**: Line-by-line comparison, not proper diff library
6. **No config for omo-switch itself**: All paths are compile-time constants

## File Reference

| Task | File | Lines |
|------|------|-------|
| Add CLI command | internal/cli/handler.go | 172 |
| Add TUI view | internal/tui/app.go | 653 |
| Add TUI component | internal/tui/components/ | 9 files |
| Add config group | internal/domain/group.go | 47 |
| Change validation | internal/domain/schema.go | 38 |
| Modify service | internal/application/service.go | 184 |
| Change config paths | internal/infrastructure/filesystem.go | 120 |
| Change backup behavior | internal/infrastructure/backup.go | 158 |

## Quick Start for AI Agents

1. **Read this file first** - understand the architecture
2. **Check AGENTS.md in target package** - if exists, read it
3. **Find similar existing code** - copy patterns, not invent new ones
4. **Run tests** - `go test ./...` before and after changes
5. **Verify boundaries** - don't cross package dependency lines

## Related Documentation

- [ai-workflow.md](.agents/rules/ai-workflow.md) - Development workflow
- [cli-architecture.md](.agents/rules/cli-architecture.md) - CLI architecture details
- [command-guidelines.md](.agents/rules/command-guidelines.md) - Command creation guide
- [go-standards.md](.agents/rules/go-standards.md) - Go engineering standards
- [testing-guide.md](.agents/rules/testing-guide.md) - Testing standards
- [bugfix-workflow.md](.agents/rules/bugfix-workflow.md) - Bug investigation
- [refactor-workflow.md](.agents/rules/refactor-workflow.md) - Refactoring guide
- [contributing-guide.md](.agents/rules/contributing-guide.md) - Developer onboarding
- [provider-integration-guide.md](.agents/rules/provider-integration-guide.md) - AI provider integration

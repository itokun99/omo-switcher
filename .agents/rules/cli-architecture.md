# CLI Architecture Guide - omo-switch

## Architecture Overview

omo-switch is a CLI/TUI application for switching oh-my-openagent configs. It uses a clean architecture pattern with clear separation of concerns.

## Package Structure

```
cmd/omo-switch/main.go          # Entry point - dependency injection
internal/
├── domain/                     # Pure business logic - NO I/O
│   ├── config.go              # Config struct + Validate()
│   ├── group.go               # Group struct + KnownGroups
│   └── schema.go              # SchemaValidator interface
├── application/                # Orchestration layer
│   └── service.go             # ConfigService
├── infrastructure/             # I/O implementations
│   ├── filesystem.go          # Store interface + FilesystemStore
│   └── backup.go              # BackupManager interface + FilesystemBackupManager
├── cli/                        # CLI mode handler
│   └── handler.go             # Handle() + cmd* functions
└── tui/                        # Bubble Tea TUI
    ├── app.go                 # App model (tea.Model)
    ├── keys.go                # Key bindings
    ├── styles.go              # Lipgloss styles
    └── components/            # 9 leaf components
        ├── list.go            # Config list
        ├── search.go          # Search filter
        ├── detail.go          # Config detail
        ├── diff.go            # Diff viewer
        ├── backup.go          # Backup manager
        ├── validate.go        # Validation results
        ├── help.go            # Help overlay
        ├── info.go            # Config info
        └── status.go          # Status bar
```

## Dependency Flow

```
main.go
├── cli.Handle(service, args, w)
│   └── service.ListConfigs(), service.SwitchConfig(), etc.
│       ├── store.ListConfigs(), store.ReadConfig()
│       └── backup.CreateBackup(), backup.ListBackups()
└── tui.NewApp(service)
    └── service.ListConfigs(), service.SwitchConfig(), etc.
        ├── store.ListConfigs(), store.ReadConfig()
        └── backup.CreateBackup(), backup.ListBackups()
```

## Entry Point (main.go)

```go
func main() {
    args := os.Args[1:]

    // Manual dependency injection
    store := infrastructure.NewFilesystemStore()
    backup := infrastructure.NewFilesystemBackupManager()
    validator := domain.DefaultValidator{}
    service := application.NewConfigService(store, backup, validator)

    // Route: no args → TUI, --cli → CLI
    if len(args) == 0 {
        app := tui.NewApp(service)
        p := tea.NewProgram(app, tea.WithAltScreen())
        p.Run()
        return
    }

    if args[0] == "--cli" {
        args = args[1:]
    }

    exitCode := cli.Handle(service, args, os.Stdout)
    os.Exit(exitCode)
}
```

## CLI Mode (internal/cli/handler.go)

### Command Dispatch

```go
func Handle(service *application.ConfigService, args []string, w io.Writer) int {
    if len(args) == 0 {
        return cmdList(service, w)
    }

    switch args[0] {
    case "--list", "-l":
        return cmdList(service, w)
    case "--current", "-c":
        return cmdCurrent(service, w)
    case "--help", "-h":
        return cmdHelp(w)
    case "show":
        if len(args) < 2 {
            fmt.Fprintln(w, "Error: show requires an alias argument")
            return 1
        }
        return cmdShow(service, args[1], w)
    default:
        return cmdSwitch(service, args[0], w)
    }
}
```

### Command Function Pattern

```go
func cmdXxx(service *application.ConfigService, w io.Writer) int {
    // 1. Call service method
    result, err := service.Xxx()
    if err != nil {
        fmt.Fprintf(w, "Error: %v\n", err)
        return 1
    }

    // 2. Format output
    fmt.Fprintf(w, "Result: %v\n", result)
    return 0
}
```

## TUI Mode (internal/tui/)

### Bubble Tea Architecture

The TUI uses the Elm architecture:
- **Model**: `App` struct (state)
- **Update**: `App.Update(msg)` (state transitions)
- **View**: `App.View()` (rendering)

### View Modes

```go
type ViewMode int

const (
    ViewList ViewMode = iota     // Default: config list
    ViewDetail                   // Config detail view
    ViewSearch                   // Search filter
    ViewHelp                     // Help overlay
    ViewBackup                   // Backup manager
    ViewDiff                     // Diff viewer
    ViewValidate                 // Validation results
    ViewInfo                     // Config info
)
```

### Message Types

```go
type configsLoadedMsg struct {
    groups   []domain.Group
    active   string
    reloaded bool
}

type errMsg struct {
    err error
}

type switchCompleteMsg struct {
    alias string
    err   error
}

// ... more message types
```

### Component Pattern

Every component in `internal/tui/components/` follows:

```go
type FooModel struct {
    // State fields
    active bool
    // ...
}

type FooStyles struct {
    // Lipgloss styles
    Header lipgloss.Style
    // ...
}

func NewFooModel() FooModel {
    return FooModel{}
}

func (m *FooModel) Render(styles FooStyles, width int) string {
    // Render component
}

func (m *FooModel) Show(...) {
    m.active = true
}

func (m *FooModel) Hide() {
    m.active = false
}

func (m FooModel) IsActive() bool {
    return m.active
}
```

**CRITICAL**: Each component owns its own `Styles` struct to avoid circular imports.

## Interfaces

### Store (infrastructure/filesystem.go)

```go
type Store interface {
    ListConfigs() (map[string]string, error)
    GetConfig(alias string) (string, error)
    ReadConfig(alias string) ([]byte, error)
    WriteConfig(alias string, content []byte) error
    ConfigDir() string
    TargetPath() string
}
```

### BackupManager (infrastructure/backup.go)

```go
type BackupManager interface {
    CreateBackup() (string, error)
    ListBackups() ([]BackupInfo, error)
    RestoreBackup(timestamp string) error
}
```

### SchemaValidator (domain/schema.go)

```go
type SchemaValidator interface {
    Validate(content []byte) error
    RequiredKeys() []string
}
```

## Config File Paths

| Path | Purpose |
|------|---------|
| `~/.config/opencode/omo_configs/omo-*.json` | Config files (switching candidates) |
| `~/.config/opencode/oh-my-openagent.json` | Active config (target) |
| `~/.config/omo-switch/backups/oh-my-openagent.<timestamp>.json` | Backups |

## Key Bindings (TUI)

| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| g/home | Jump to top |
| G/end | Jump to bottom |
| Enter | Switch config |
| s | Show detail |
| / | Search |
| ? | Help |
| v | Validate all |
| b | Backup manager |
| d | Diff viewer |
| i | Config info |
| r | Reload |
| q/Esc | Quit |

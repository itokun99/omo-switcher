# Command Guidelines - omo-switch

## Overview

This document defines standards for creating CLI commands in omo-switch.

## Command Architecture

omo-switch uses **manual dispatch** - no Cobra, no urfave/cli. Commands are defined in `internal/cli/handler.go` using a switch statement.

## Adding a New Command

### Step 1: Add Case in Handle()

```go
func Handle(service *application.ConfigService, args []string, w io.Writer) int {
    switch args[0] {
    case "--list", "-l":
        return cmdList(service, w)
    case "--mycommand", "-m":  // Add here
        return cmdMyCommand(service, w)
    // ...
    }
}
```

### Step 2: Create Command Function

```go
func cmdMyCommand(service *application.ConfigService, w io.Writer) int {
    // 1. Validate input (if needed)
    // 2. Call service method
    // 3. Format output
    // 4. Return exit code
}
```

### Step 3: Follow Naming Convention

| Element | Convention | Example |
|---------|------------|---------|
| Function name | `cmd` + CamelCase | `cmdMyCommand` |
| Long flag | `--kebab-case` | `--my-command` |
| Short flag | `-letter` | `-m` |
| Exit code | 0 success, 1 error | `return 0` |

## Command Function Template

```go
func cmdXxx(service *application.ConfigService, w io.Writer) int {
    // Call service
    result, err := service.Xxx()
    if err != nil {
        fmt.Fprintf(w, "Error: %v\n", err)
        return 1
    }

    // Format output
    fmt.Fprintf(w, "Result: %v\n", result)
    return 0
}
```

## Error Handling

```go
// Standard error pattern
if err != nil {
    fmt.Fprintf(w, "Error: %v\n", err)
    return 1
}

// Unknown flag detection
if strings.HasPrefix(args[0], "-") {
    fmt.Fprintf(w, "Error: unknown option %s\n", args[0])
    fmt.Fprintln(w, "Run 'omo-switch --help' for usage.")
    return 1
}
```

## Output Formatting

### List Output

```go
fmt.Fprint(w, "\nAvailable configs:\n\n")
for _, group := range groups {
    fmt.Fprintf(w, "  %s\n", group.Name)
    for _, cfg := range group.Configs {
        fmt.Fprintf(w, "    %-20s → %s\n", cfg.Alias, cfg.FileName)
    }
}
```

### Current Output

```go
fmt.Fprintf(w, "Current: %s\n", alias)
// or
fmt.Fprintln(w, "No active config matched")
```

### Show Output

```go
fmt.Fprintf(w, "--- %s (%s) ---\n%s\n", alias, fileName, content)
```

### Switch Output

```go
fmt.Fprintf(w, "Switched to: %s\n", alias)
```

## Help Text

Update `cmdHelp()` when adding new commands:

```go
func cmdHelp(w io.Writer) int {
    fmt.Fprint(w, `
Usage: omo-switch [command] [alias]

Commands:
  (no args)          List all available configs
  --list, -l         List all available configs
  --current, -c      Show active config alias
  --mycommand, -m    My new command description
  show <alias>       Print content of a config
  <alias>            Switch to that config

Examples:
  omo-switch                  # list all
  omo-switch claude           # switch to claude config
  omo-switch show optimized-high
  omo-switch --current
`)
    return 0
}
```

## Testing Commands

### Test Structure

```go
func TestHandleMyCommand(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        wantCode   int
        wantOutput []string
    {
        {
            name:       "success case",
            args:       []string{"--mycommand"},
            wantCode:   0,
            wantOutput: []string{"expected output"},
        },
        {
            name:       "error case",
            args:       []string{"--mycommand", "invalid"},
            wantCode:   1,
            wantOutput: []string{"Error:"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock
            mock := &mockStore{...}

            // Create service
            service := newTestService(mock, nil)

            // Capture output
            var buf bytes.Buffer
            got := Handle(service, tt.args, &buf)

            // Assert exit code
            if got != tt.wantCode {
                t.Errorf("Handle() = %d, want %d", got, tt.wantCode)
            }

            // Assert output
            output := buf.String()
            for _, want := range tt.wantOutput {
                if !strings.Contains(output, want) {
                    t.Errorf("output missing %q", want)
                }
            }
        })
    }
}
```

## Command Reference

| Command | Flag | Function | Description |
|---------|------|----------|-------------|
| List | `--list`, `-l` | `cmdList()` | List all configs |
| Current | `--current`, `-c` | `cmdCurrent()` | Show active config |
| Help | `--help`, `-h` | `cmdHelp()` | Show usage |
| Show | `show <alias>` | `cmdShow()` | Print config content |
| Switch | `<alias>` | `cmdSwitch()` | Switch to config |

## Adding Subcommands

To add a subcommand (e.g., `omo-switch backup list`):

```go
case "backup":
    if len(args) < 2 {
        fmt.Fprintln(w, "Error: backup requires subcommand")
        return 1
    }
    switch args[1] {
    case "list":
        return cmdBackupList(service, w)
    case "restore":
        return cmdBackupRestore(service, args[2:], w)
    default:
        fmt.Fprintf(w, "Error: unknown backup subcommand %s\n", args[1])
        return 1
    }
}
```

## Common Patterns

### Argument Validation

```go
if len(args) < 2 {
    fmt.Fprintln(w, "Error: show requires an alias argument")
    return 1
}
```

### Flag Detection

```go
if strings.HasPrefix(args[0], "-") {
    fmt.Fprintf(w, "Error: unknown option %s\n", args[0])
    return 1
}
```

### Service Error Handling

```go
result, err := service.Xxx()
if err != nil {
    fmt.Fprintf(w, "Error: %v\n", err)
    return 1
}
```

# Bug Fix Workflow - omo-switch

## Overview

This document defines the systematic approach to investigating and fixing bugs in omo-switch.

## Bug Fix Process

### Phase 1: Reproduce

```
1. Understand the bug report
   - What happened?
   - What was expected?
   - Steps to reproduce?

2. Create minimal reproduction
   - Use CLI mode for easier debugging
   - Isolate the issue to specific command/feature

3. Verify the bug exists
   - Run reproduction steps
   - Confirm unexpected behavior
```

### Phase 2: Investigate

```
1. Trace execution path
   - Start from entry point (main.go)
   - Follow the call chain
   - Identify where behavior diverges

2. Check package boundaries
   - Is the issue in domain logic?
   - Is it in infrastructure I/O?
   - Is it in application orchestration?
   - Is it in CLI/TUI presentation?

3. Identify root cause
   - Don't just find symptoms
   - Find the actual source of the problem
```

### Phase 3: Fix

```
1. Create fix proposal
   - What needs to change?
   - What files are affected?
   - What are the side effects?

2. Implement minimal fix
   - Change only what's necessary
   - Don't refactor while fixing
   - Follow existing patterns

3. Write regression test
   - Test that reproduces the bug
   - Verify fix prevents recurrence
```

### Phase 4: Verify

```
1. Run affected tests
   go test ./internal/affected_package/...

2. Run all tests
   go test ./...

3. Manual verification
   - Test the original reproduction steps
   - Verify fix works as expected
```

## Debugging Techniques

### CLI Debugging

```bash
# Run specific command
./omo-switch --list

# Run with verbose output (if implemented)
./omo-switch --verbose --list

# Check exit code
echo $?
```

### TUI Debugging

TUI is harder to debug. Use CLI mode when possible:

```bash
# Force CLI mode
./omo-switch --cli --list

# Or use specific commands
./omo-switch show claude
```

### Adding Debug Output

Temporary debug output (remove before commit):

```go
fmt.Fprintf(os.Stderr, "DEBUG: variable = %v\n", variable)
```

### Using Tests for Debugging

```go
func TestDebugIssue(t *testing.T) {
    // Setup specific scenario
    mock := &mockStore{
        configs: map[string]string{"test": "omo-test.json"},
        content: map[string][]byte{"test": []byte(`{"invalid": true}`)},
    }

    service := newTestService(mock, nil)

    // Test the specific operation
    _, err := service.ListConfigs()
    if err != nil {
        t.Logf("Error: %v", err)
    }
}
```

## Common Bug Categories

### Category 1: Config File Issues

**Symptoms**: Configs not found, invalid configs, wrong active config

**Investigation**:
```go
// Check config directory
store := infrastructure.NewFilesystemStore()
configs, err := store.ListConfigs()
fmt.Printf("Configs: %v, Error: %v\n", configs, err)

// Check specific config
content, err := store.ReadConfig("claude")
fmt.Printf("Content: %s, Error: %v\n", content, err)
```

**Common causes**:
- Wrong file path
- Invalid JSON
- Missing `agents` key
- File permissions

### Category 2: CLI Command Issues

**Symptoms**: Wrong output, missing output, wrong exit code

**Investigation**:
```go
// Test command directly
var buf bytes.Buffer
code := cli.Handle(service, []string{"--list"}, &buf)
fmt.Printf("Exit: %d, Output: %s\n", code, buf.String())
```

**Common causes**:
- Wrong argument parsing
- Missing error handling
- Wrong output format

### Category 3: TUI Issues

**Symptoms**: UI not rendering, wrong behavior, crashes

**Investigation**:
- Use CLI mode to isolate
- Check ViewMode transitions
- Check message handling

**Common causes**:
- Wrong message type handling
- Missing state updates
- Wrong component lifecycle

### Category 4: Infrastructure Issues

**Symptoms**: File not found, permission errors, wrong paths

**Investigation**:
```go
// Check paths
store := infrastructure.NewFilesystemStore()
fmt.Printf("ConfigDir: %s\n", store.ConfigDir())
fmt.Printf("TargetPath: %s\n", store.TargetPath())

// Check file existence
_, err := os.Stat(store.TargetPath())
fmt.Printf("Target exists: %v\n", err == nil)
```

**Common causes**:
- Wrong path construction
- Missing directory creation
- Permission issues

## Bug Fix Checklist

Before submitting fix:

- [ ] Bug reproduced and understood
- [ ] Root cause identified (not just symptom)
- [ ] Minimal fix implemented
- [ ] Regression test written
- [ ] All tests pass (`go test ./...`)
- [ ] No package boundary violations
- [ ] Existing patterns followed
- [ ] No unnecessary refactoring

## Example Bug Fix

### Bug Report
"omo-switch shows wrong active config"

### Investigation

```go
// 1. Check how active config is determined
func (s *ConfigService) GetActiveConfig() (string, error) {
    targetContent, err := os.ReadFile(s.store.TargetPath())
    // ...

    for alias := range aliases {
        content, err := s.store.ReadConfig(alias)
        // ...
        if string(content) == string(targetContent) {
            return alias, nil
        }
    }
    return "", nil
}

// 2. Issue: byte comparison is sensitive to whitespace
// 3. Fix: normalize JSON before comparison
```

### Fix

```go
// Normalize JSON before comparison
func normalizeJSON(data []byte) ([]byte, error) {
    var parsed interface{}
    if err := json.Unmarshal(data, &parsed); err != nil {
        return nil, err
    }
    return json.Marshal(parsed)
}
```

### Regression Test

```go
func TestGetActiveConfig_WhitespaceInsensitive(t *testing.T) {
    // Test with different whitespace in same JSON
    target := []byte(`{"agents": {}}`)
    config := []byte(`{
        "agents": {}
    }`)

    // Should match despite whitespace differences
}
```

## Debugging Tools

### go vet

```bash
go vet ./...
```

### Race Detector

```bash
go test -race ./...
```

### Verbose Tests

```bash
go test -v ./internal/domain/...
```

### Specific Test

```bash
go test -run TestGetActiveConfig ./internal/application/...
```

# Refactoring Workflow - omo-switch

## Overview

This document defines the safe refactoring approach for omo-switch.

## Refactoring Principles

1. **Behavior Preservation**: Refactoring must not change external behavior
2. **Incremental Steps**: Small, verifiable changes
3. **Test Coverage**: Tests must pass before and after each step
4. **Pattern Consistency**: Follow existing codebase patterns

## Refactoring Process

### Phase 1: Analyze

```
1. Understand current implementation
   - Read the code thoroughly
   - Understand all dependencies
   - Note all callers/users

2. Identify refactoring goal
   - What's wrong with current design?
   - What should it look like after?
   - Is refactoring necessary?

3. Assess risk
   - How many files affected?
   - How many tests affected?
   - What could break?
```

### Phase 2: Plan

```
1. Create step-by-step plan
   - Each step should be verifiable
   - Each step should leave code working
   - Order steps to minimize risk

2. Identify test requirements
   - What tests need updating?
   - What new tests needed?
   - What tests verify behavior preservation?

3. Estimate effort
   - Simple refactor: 1-2 files
   - Medium refactor: 3-5 files
   - Large refactor: 6+ files (consider breaking up)
```

### Phase 3: Execute

```
1. Ensure tests pass initially
   go test ./...

2. Make one small change
   - Extract function
   - Rename variable
   - Move code
   - etc.

3. Run tests after each change
   go test ./...

4. Repeat until complete
```

### Phase 4: Verify

```
1. Run all tests
   go test ./...

2. Run vet
   go vet ./...

3. Manual verification
   - Test affected features
   - Verify no behavior changes
```

## Safe Refactoring Patterns

### Pattern 1: Extract Function

**Before**:
```go
func (s *ConfigService) ListConfigs() ([]domain.Group, error) {
    aliases, err := s.store.ListConfigs()
    if err != nil {
        return nil, fmt.Errorf("listing configs: %w", err)
    }

    groups := make(map[string]*domain.Group)
    for _, name := range knownGroupNames() {
        g := domain.NewGroup(name)
        groups[name] = &g
    }

    for alias, filename := range aliases {
        content, err := s.store.ReadConfig(alias)
        if err != nil {
            slog.Error("reading config", "alias", alias, "error", err)
            continue
        }
        filePath := filepath.Join(s.store.ConfigDir(), filename)
        cfg := domain.NewConfig(alias, filename, filePath, content)
        cfg = cfg.Validate(s.validator)
        groupName := domain.GetGroupForAlias(alias)
        groups[groupName].AddConfig(cfg)
    }

    result := make([]domain.Group, 0, len(groups))
    for _, name := range knownGroupNames() {
        result = append(result, *groups[name])
    }
    return result, nil
}
```

**After**:
```go
func (s *ConfigService) ListConfigs() ([]domain.Group, error) {
    aliases, err := s.store.ListConfigs()
    if err != nil {
        return nil, fmt.Errorf("listing configs: %w", err)
    }

    groups := s.initGroups()
    s.populateGroups(groups, aliases)
    return s.collectGroups(groups), nil
}

func (s *ConfigService) initGroups() map[string]*domain.Group {
    groups := make(map[string]*domain.Group)
    for _, name := range knownGroupNames() {
        g := domain.NewGroup(name)
        groups[name] = &g
    }
    return groups
}

func (s *ConfigService) populateGroups(groups map[string]*domain.Group, aliases map[string]string) {
    for alias, filename := range aliases {
        content, err := s.store.ReadConfig(alias)
        if err != nil {
            slog.Error("reading config", "alias", alias, "error", err)
            continue
        }
        filePath := filepath.Join(s.store.ConfigDir(), filename)
        cfg := domain.NewConfig(alias, filename, filePath, content)
        cfg = cfg.Validate(s.validator)
        groupName := domain.GetGroupForAlias(alias)
        groups[groupName].AddConfig(cfg)
    }
}

func (s *ConfigService) collectGroups(groups map[string]*domain.Group) []domain.Group {
    result := make([]domain.Group, 0, len(groups))
    for _, name := range knownGroupNames() {
        result = append(result, *groups[name])
    }
    return result
}
```

### Pattern 2: Extract Interface

**Before**:
```go
type ConfigService struct {
    store     *infrastructure.FilesystemStore
    backup    *infrastructure.FilesystemBackupManager
    validator domain.SchemaValidator
}
```

**After**:
```go
type ConfigService struct {
    store     infrastructure.Store
    backup    infrastructure.BackupManager
    validator domain.SchemaValidator
}
```

### Pattern 3: Move Function to Different Package

**Before** (in application/service.go):
```go
func knownGroupNames() []string {
    return []string{"Mono", "Optimized", "Low-Cost", "Custom"}
}
```

**After** (in domain/group.go):
```go
func KnownGroupNames() []string {
    return []string{"Mono", "Optimized", "Low-Cost", "Custom"}
}
```

### Pattern 4: Rename for Clarity

**Before**:
```go
func (s *ConfigService) GetActiveConfig() (string, error) {
```

**After**:
```go
func (s *ConfigService) FindActiveConfigAlias() (string, error) {
```

## Refactoring Checklist

Before starting:

- [ ] All tests pass
- [ ] Understand all callers/users
- [ ] Have step-by-step plan
- [ ] Know what tests verify behavior

During refactoring:

- [ ] One small change at a time
- [ ] Run tests after each change
- [ ] Don't change behavior
- [ ] Follow existing patterns

After refactoring:

- [ ] All tests pass
- [ ] No vet warnings
- [ ] Build succeeds
- [ ] Behavior preserved
- [ ] Code is cleaner

## Common Refactoring Scenarios

### Scenario 1: Duplicate Code

**Problem**: Same logic in multiple places

**Solution**: Extract to shared function

**Risk**: Low (if tests exist for both locations)

**Example**:
```go
// Before: duplicate path construction
func (s *FilesystemStore) ReadConfig(alias string) ([]byte, error) {
    path := filepath.Join(s.configDir, fmt.Sprintf("omo-%s.json", alias))
    // ...
}

func (s *FilesystemStore) WriteConfig(alias string, content []byte) error {
    path := filepath.Join(s.configDir, fmt.Sprintf("omo-%s.json", alias))
    // ...
}

// After: shared helper
func (s *FilesystemStore) configPath(alias string) string {
    return filepath.Join(s.configDir, fmt.Sprintf("omo-%s.json", alias))
}
```

### Scenario 2: Long Function

**Problem**: Function does too many things

**Solution**: Extract sub-functions

**Risk**: Medium (need to verify extracted functions work correctly)

**Example**:
```go
// Before: 50-line function
func (s *ConfigService) ListConfigs() ([]domain.Group, error) {
    // ... 50 lines
}

// After: composed of smaller functions
func (s *ConfigService) ListConfigs() ([]domain.Group, error) {
    aliases := s.discoverConfigs()
    groups := s.initGroups()
    s.populateGroups(groups, aliases)
    return s.collectGroups(groups), nil
}
```

### Scenario 3: Interface Extraction

**Problem**: Concrete type used where interface would be better

**Solution**: Define interface, change parameter type

**Risk**: Low (compile-time checks catch errors)

**Example**:
```go
// Before
func NewConfigService(store *FilesystemStore, ...) *ConfigService {

// After
type Store interface {
    ListConfigs() (map[string]string, error)
    // ...
}

func NewConfigService(store Store, ...) *ConfigService {
```

### Scenario 4: Move Code to Better Package

**Problem**: Code in wrong package

**Solution**: Move to correct package, update imports

**Risk**: Medium (need to update all callers)

**Example**:
```go
// Before: in application/service.go
func knownGroupNames() []string {
    return []string{"Mono", "Optimized", "Low-Cost", "Custom"}
}

// After: in domain/group.go
func KnownGroupNames() []string {
    return []string{"Mono", "Optimized", "Low-Cost", "Custom"}
}
```

## Refactoring Anti-Patterns

### Anti-Pattern 1: Refactor + Feature

**Wrong**: Add new feature while refactoring

**Right**: Refactor first, then add feature

### Anti-Pattern 2: Big Bang Refactor

**Wrong**: Change everything at once

**Right**: Small, incremental changes

### Anti-Pattern 3: Skip Tests

**Wrong**: Refactor without running tests

**Right**: Run tests after every change

### Anti-Pattern 4: Change Behavior

**Wrong**: "Improve" behavior while refactoring

**Right**: Preserve exact behavior

## Large Refactoring Strategy

For refactoring that touches 6+ files:

1. **Break into smaller refactorings**
   - Each should be independently valuable
   - Each should be independently verifiable

2. **Create feature branch**
   - Isolate refactoring from other work
   - Easy to abandon if problems arise

3. **Commit frequently**
   - Each commit should be working state
   - Easy to revert specific changes

4. **Review carefully**
   - Have someone review the refactoring
   - Or use AI review tools

## Refactoring Tools

### go vet

```bash
go vet ./...
```

### Tests

```bash
go test ./...
go test -race ./...
```

### Build

```bash
go build ./cmd/omo-switch
```

### Manual Testing

Test affected features after refactoring.

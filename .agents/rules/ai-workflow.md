# AI Workflow - Development Guidelines for omo-switch

## Overview

This document defines the AI-assisted development workflow for the omo-switch project. It ensures consistent, high-quality contributions from any AI agent.

## Pre-Development Checklist

Before writing ANY code:

1. **Read AGENTS.md** - Understand architecture boundaries
2. **Read this file** - Understand workflow requirements
3. **Find similar code** - Copy existing patterns
4. **Run tests** - `go test ./...` to establish baseline

## Development Workflow

### Phase 1: Understanding

```
1. Identify the task type:
   - New CLI command → Read COMMAND_GUIDELINES.md
   - New TUI component → Read CLI_ARCHITECTURE.md
   - Bug fix → Read BUGFIX_WORKFLOW.md
   - Refactor → Read REFACTOR_WORKFLOW.md
   - New test → Read TESTING_GUIDE.md

2. Find reference implementation:
   - Search for similar existing code
   - Understand the pattern used
   - Note the package boundaries

3. Identify affected files:
   - Which packages need changes?
   - What interfaces are involved?
   - What tests need updating?
```

### Phase 2: Implementation

```
1. Create todo list with specific steps
2. Implement one step at a time
3. Run tests after each step
4. Verify package boundaries not violated
```

### Phase 3: Verification

```
1. Run `go test ./...` - all tests must pass
2. Run `go vet ./...` - no warnings
3. Run `go build ./cmd/omo-switch` - compiles successfully
4. Manual verification if applicable
```

## Task-Specific Workflows

### Adding a New CLI Command

```
1. Read COMMAND_GUIDELINES.md
2. Add case in internal/cli/handler.go Handle()
3. Create cmdXxx() function following existing pattern
4. Add tests in handler_test.go
5. Update help text in cmdHelp()
6. Run: go test ./internal/cli/...
```

### Adding a New TUI Component

```
1. Read CLI_ARCHITECTURE.md (TUI section)
2. Create component in internal/tui/components/
3. Follow FooModel/FooStyles pattern
4. Add ViewMode constant in app.go
5. Add key binding in keys.go
6. Add handler in app.go handleKey()
7. Add render logic in View()
8. Create test file with table-driven tests
9. Run: go test ./internal/tui/...
```

### Adding a New Config Group

```
1. Edit internal/domain/group.go
2. Add entry to KnownGroups map
3. Update tests in group_test.go
4. Run: go test ./internal/domain/...
```

### Modifying Validation Logic

```
1. Edit internal/domain/schema.go
2. Update DefaultValidator.Validate()
3. Update RequiredKeys() if needed
4. Update tests in schema_test.go
5. Run: go test ./internal/domain/...
```

## Error Recovery

If you encounter issues:

1. **Don't guess** - Read the actual code
2. **Don't assume** - Verify with tests
3. **Don't skip** - Complete all steps
4. **Don't break** - Maintain package boundaries

## Quality Gates

Before marking task complete:

- [ ] All tests pass (`go test ./...`)
- [ ] No vet warnings (`go vet ./...`)
- [ ] Build succeeds (`go build ./cmd/omo-switch`)
- [ ] Package boundaries respected
- [ ] Existing patterns followed
- [ ] Tests added/updated for changes

## Common Mistakes to Avoid

| Mistake | Why It's Wrong | How to Avoid |
|---------|---------------|--------------|
| Adding I/O to domain | Breaks architecture | Check AGENTS.md Rule 1 |
| Using Cobra | Project uses manual dispatch | Check AGENTS.md Rule 3 |
| Sharing styles | Causes circular imports | Check AGENTS.md Rule 4 |
| External test frameworks | Inconsistent conventions | Check AGENTS.md Rule 5 |
| Creating new packages | Over-engineering | Use existing packages |
| Skipping tests | Quality regression | Always write tests |

## Reference Files

| File | Purpose |
|------|---------|
| AGENTS.md | Master rules |
| CLI_ARCHITECTURE.md | Architecture details |
| COMMAND_GUIDELINES.md | Command creation |
| GO_STANDARDS.md | Go conventions |
| TESTING_GUIDE.md | Test patterns |
| BUGFIX_WORKFLOW.md | Bug investigation |
| REFACTOR_WORKFLOW.md | Refactoring guide |
| CONTRIBUTING_GUIDE.md | Onboarding |

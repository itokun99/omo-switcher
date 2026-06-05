# AI Agent Rules - omo-switch

This directory contains documentation for AI agents working on the omo-switch project.

## Files

| File | Purpose |
|------|---------|
| [ai-workflow.md](ai-workflow.md) | Development workflow for AI-assisted development |
| [cli-architecture.md](cli-architecture.md) | CLI/TUI architecture details |
| [command-guidelines.md](command-guidelines.md) | Standards for creating CLI commands |
| [go-standards.md](go-standards.md) | Go engineering standards |
| [testing-guide.md](testing-guide.md) | Testing patterns and conventions |
| [bugfix-workflow.md](bugfix-workflow.md) | Bug investigation and fixing workflow |
| [refactor-workflow.md](refactor-workflow.md) | Safe refactoring approach |
| [contributing-guide.md](contributing-guide.md) | Developer onboarding guide |
| [provider-integration-guide.md](provider-integration-guide.md) | AI provider integration standards |

## Quick Start

1. Read [../AGENTS.md](../AGENTS.md) for master rules
2. Read [contributing-guide.md](contributing-guide.md) for getting started
3. Read the relevant guide for your task

## Architecture Summary

```
cmd/omo-switch/main.go          # Entry point
internal/
├── domain/                     # Pure business logic (no I/O)
├── application/                # ConfigService orchestrator
├── infrastructure/             # Filesystem I/O
├── cli/                        # CLI handler
└── tui/                        # Bubble Tea TUI
    └── components/             # UI components
```

## Key Rules

1. **Package boundaries**: domain → application → infrastructure → cli/tui
2. **No I/O in domain**: Pure business logic only
3. **Interface-based DI**: All infrastructure via interfaces
4. **Manual dispatch**: No Cobra, no CLI frameworks
5. **Table-driven tests**: Standard testing package only
6. **Error wrapping**: `fmt.Errorf("context: %w", err)`

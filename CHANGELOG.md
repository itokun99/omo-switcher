# Changelog

## [2.0.0] - 2026-06-04

### Added
- Interactive TUI mode (Bubble Tea) as default
- Search/filter configs with `/` key
- Config detail view with `s` key
- Help overlay with `?` key
- Config validation with `v` key
- Backup manager with `b` key
- Config diff viewer with `d` key
- Config info display with `i` key
- Reload configs with `r` key
- Cross-platform builds (macOS, Linux, Windows)
- GitHub Actions release workflow

### Changed
- Rewritten from Node.js to Go
- TUI mode is now default (no args)
- CLI mode available via `--cli` flag
- Layered architecture (domain/application/infrastructure/tui)

### Preserved
- All CLI commands (--list, --current, show, alias)
- Config discovery from ~/.config/opencode/omo_configs/
- Schema validation before switching
- Auto-backup before switching
- Grouped display (Mono, Optimized, Low-Cost, Custom)

## [1.0.0] - 2025-01-01

### Added
- Initial Node.js CLI implementation
- Config discovery and switching
- Schema validation
- Auto-backup

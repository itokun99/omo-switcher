# TUI Components

9 Bubble Tea components. Each owns its styles struct.

## PATTERN
```go
type FooModel struct { /* state */ }
type FooStyles struct { /* lipgloss styles */ }
func NewFooModel(...) FooModel
func (m *FooModel) Render(styles FooStyles, width int) string
func (m *FooModel) Show(...)
func (m *FooModel) Hide()
func (m FooModel) IsActive() bool
```

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Modify config list | `list.go` | `ListModel` with cursor, scroll, grouping |
| Add search filter | `search.go` | `SearchModel` with `textinput` |
| Change detail view | `detail.go` | `DetailModel` with scroll |
| Add help overlay | `help.go` | `HelpModel` toggled by `?` |
| Modify validation | `validate.go` | `ValidateModel` shows bulk results |
| Change backup UI | `backup.go` | `BackupModel` with restore confirmation |
| Add diff view | `diff.go` | `DiffModel` side-by-side comparison |
| Show config info | `info.go` | `InfoModel` metadata display |
| Change status bar | `status.go` | `StatusModel` bottom bar |

## CONVENTIONS
- Each component defines its own `Styles` struct — never share styles across components
- Components are stateful: `Show()` activates, `Hide()` deactivates, `IsActive()` checks
- `Render()` takes styles + width, returns string
- Tests: table-driven with `t.Run()`, mock service via hand-written struct

## ANTI-PATTERNS
- Do NOT put styles in a shared location — causes circular imports
- Do NOT access `tui.App` from components — components are leaf nodes

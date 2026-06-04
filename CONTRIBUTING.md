# Contributing to omo-switch

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/itokun99/omo-switch.git
   cd omo-switch
   ```

2. Install Go 1.22+

3. Build:
   ```bash
   go build -o omo-switch ./cmd/omo-switch
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

## Project Structure

```
├── cmd/omo-switch/     # Entry point
├── internal/
│   ├── domain/         # Business models (Config, Group, Schema)
│   ├── application/    # Service layer (ConfigService)
│   ├── infrastructure/ # I/O (filesystem, backup)
│   ├── cli/            # CLI handler
│   └── tui/            # Bubble Tea TUI
│       └── components/ # TUI components
├── scripts/            # Build scripts
└── .github/workflows/  # CI/CD
```

## Code Guidelines

- Follow Go conventions (gofmt, go vet)
- Table-driven tests with `t.Run()`
- Domain layer: no I/O dependencies
- Infrastructure layer: no business logic
- TUI components: own styles struct to avoid circular imports

## Pull Requests

1. Fork the repo
2. Create a feature branch
3. Write tests for new functionality
4. Ensure `go test ./...` passes
5. Submit PR with clear description

## License

MIT

# @indrawandev/omo-switch

CLI/TUI switcher for [oh-my-openagent](https://github.com/code-yeongyu/oh-my-openagent) configs.

Instantly swap between different model configurations for opencode without manually copying files.

## Installation

### Go Binary (Recommended)

```bash
go install github.com/itokun99/omo-switch/cmd/omo-switch@latest
```

### Homebrew

```bash
brew tap itokun99/omo-switch
brew install omo-switch
```

### From Source

```bash
git clone https://github.com/itokun99/omo-switch.git
cd omo-switch
go build -o omo-switch ./cmd/omo-switch
```

### npm

```bash
npm install -g @indrawandev/omo-switch
```

### bun

```bash
bun install -g @indrawandev/omo-switch
```

## Usage

### TUI Mode (Default)

```bash
omo-switch
```

Opens an interactive terminal UI for browsing, searching, and switching configs. Use arrow keys or vim bindings to navigate. Press Enter to switch.

### CLI Mode (Backward Compatible)

```bash
omo-switch --list          # List all configs
omo-switch --current       # Show active config
omo-switch claude          # Switch to claude config
omo-switch show claude     # Show config content
omo-switch --cli --list    # Explicit CLI mode
```

## Key Bindings

| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| g/home | Jump to top |
| G/end | Jump to bottom |
| Enter | Switch to config |
| s | Show config detail |
| / | Search configs |
| ? | Toggle help |
| v | Validate all configs |
| b | Backup manager |
| d | Diff viewer |
| i | Config info |
| r | Reload configs |
| q/Esc | Quit |

## Features

- **Auto-discovery** — scans `~/.config/opencode/omo_configs/omo-*.json`, no hardcoding needed
- **Schema validation** — validates config has required `agents` key before applying
- **Auto-backup** — backs up current active config to `~/.config/omo-switch/backups/` before every switch
- **Grouped display** — configs grouped by Mono / Optimized / Low-Cost / Custom
- **Interactive TUI** — browse, search, and switch configs with keyboard navigation
- **Cross-platform** — builds for macOS, Linux, and Windows

## Configuration

Configs are stored in `~/.config/opencode/omo_configs/` as `omo-*.json` files.

The active config is at `~/.config/opencode/oh-my-openagent.json`.

Backups are stored in `~/.config/omo-switch/backups/`.

## Config Discovery

`omo-switch` auto-discovers all `omo-*.json` files from `~/.config/opencode/omo_configs/`.

Any file you add there with the `omo-` prefix will appear automatically under the **Custom** section — no config changes needed.

```
~/.config/opencode/omo_configs/
├── omo-optimized-high.json     → optimized-high
├── omo-optimized-medium.json   → optimized-medium
├── omo-optimized-low.json      → optimized-low
├── omo-lc-mode-low.json        → lc-mode-low
├── omo-lc-mode-medium.json     → lc-mode-medium
├── omo-lc-mode-high.json       → lc-mode-high
├── omo-lc-mode-ultra.json      → lc-mode-ultra
├── omo-minimax.json            → minimax
├── omo-qwen.json               → qwen
├── omo-deepseek.json           → deepseek
├── omo-glm.json                → glm
├── omo-gpt.json                → gpt
├── omo-claude.json             → claude
└── omo-my-custom.json          → my-custom  (Custom section)
```

## List Output

```
Available configs:

  Mono
    minimax              → omo-minimax.json
    qwen                 → omo-qwen.json
    deepseek             → omo-deepseek.json
    glm                  → omo-glm.json
    gpt                  → omo-gpt.json
    claude               → omo-claude.json
  Optimized
    optimized-high       → omo-optimized-high.json ◀ active
    optimized-medium     → omo-optimized-medium.json
    optimized-low        → omo-optimized-low.json
  Low-Cost
    lc-mode-low          → omo-lc-mode-low.json
    lc-mode-medium       → omo-lc-mode-medium.json
    lc-mode-high         → omo-lc-mode-high.json
    lc-mode-ultra        → omo-lc-mode-ultra.json
  Custom
    my-custom            → omo-my-custom.json
```

## Config File Format

Each config file follows the oh-my-openagent schema:

```json
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json",
  "agents": {
    "sisyphus": {
      "model": "opencode-go/kimi-k2.6"
    }
  },
  "categories": {
    "deep": {
      "model": "opencode-go/deepseek-v4-pro",
      "variant": "medium"
    }
  }
}
```

See the [oh-my-openagent agent model matching guide](https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/agent-model-matching.md) for model family recommendations per agent.

## License

MIT

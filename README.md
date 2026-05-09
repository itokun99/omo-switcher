# omo-switch

CLI switcher for [oh-my-openagent](https://github.com/code-yeongyu/oh-my-openagent) configs.

Instantly swap between different model configurations for opencode without manually copying files.

## Installation

```bash
npm install -g omo-switch
```

Or with bun:

```bash
bun install -g omo-switch
```

## Usage

```bash
# List all available configs
omo-switch --list

# Switch to a config (validates schema + auto-backup before applying)
omo-switch optimized-high

# Show currently active config
omo-switch --current

# Print content of a specific config
omo-switch show optimized-high
```

## Features

- **Auto-discovery** — scans `~/.config/opencode/omo_configs/omo-*.json`, no hardcoding needed
- **Schema validation** — validates config has required `agents` key before applying
- **Auto-backup** — backs up current active config to `~/.config/omo-switch/backups/` before every switch
- **Grouped display** — configs grouped by Mono / Optimized / Low-Cost / Custom

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

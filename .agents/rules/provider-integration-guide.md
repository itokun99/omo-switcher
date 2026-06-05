# Provider Integration Guide - omo-switch

## Overview

This document defines standards for integrating new AI providers or config types into omo-switch.

## Current Architecture

omo-switch manages config files for oh-my-openagent. Each config file:
- Lives in `~/.config/opencode/omo_configs/omo-*.json`
- Has an alias (filename without `omo-` prefix and `.json` suffix)
- Contains JSON with at least an `agents` key
- Gets classified into a group (Mono, Optimized, Low-Cost, Custom)

## Adding a New Config Type

### Step 1: Create Config File

Create a new JSON file in the config directory:

```bash
# Example: adding a new provider config
cat > ~/.config/opencode/omo_configs/omo-newprovider.json << 'EOF'
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json",
  "agents": {
    "sisyphus": {
      "model": "newprovider/model-name"
    }
  },
  "categories": {
    "deep": {
      "model": "newprovider/deep-model",
      "variant": "medium"
    }
  }
}
EOF
```

### Step 2: Verify Discovery

omo-switch auto-discovers configs:

```bash
# List all configs
./omo-switch --list

# Check if new config appears
./omo-switch show newprovider
```

### Step 3: Add to Known Group (Optional)

If the config should be in a specific group, edit `internal/domain/group.go`:

```go
var KnownGroups = map[string][]string{
    "Mono":      {"minimax", "qwen", "deepseek", "glm", "gpt", "claude"},
    "Optimized": {"optimized-high", "optimized-medium", "optimized-low"},
    "Low-Cost":  {"lc-mode-low", "lc-mode-medium", "lc-mode-high", "lc-mode-ultra"},
    "NewGroup":  {"newprovider", "newprovider-alt"},  // Add here
}
```

### Step 4: Update Tests

Update tests in `internal/domain/group_test.go`:

```go
func TestGetGroupForAlias(t *testing.T) {
    tests := []struct {
        name  string
        alias string
        want  string
    }{
        // ... existing tests
        {name: "newprovider in NewGroup", alias: "newprovider", want: "NewGroup"},
    }
    // ...
}
```

## Config File Schema

### Required Fields

```json
{
  "agents": {
    "agent-name": {
      "model": "provider/model-name"
    }
  }
}
```

### Optional Fields

```json
{
  "$schema": "https://...",
  "agents": { ... },
  "categories": {
    "category-name": {
      "model": "provider/model-name",
      "variant": "low|medium|high"
    }
  }
}
```

### Validation

The `DefaultValidator` in `internal/domain/schema.go` checks:
- JSON is valid
- `agents` key exists

To add custom validation:

```go
// In domain/schema.go
func (v DefaultValidator) Validate(content []byte) error {
    var parsed map[string]any
    if err := json.Unmarshal(content, &parsed); err != nil {
        return fmt.Errorf("invalid json: %w", err)
    }

    if _, ok := parsed["agents"]; !ok {
        return fmt.Errorf("missing required key: agents")
    }

    // Add custom validation here
    // ...

    return nil
}
```

## Adding a New Group

### Step 1: Define Group

In `internal/domain/group.go`:

```go
var KnownGroups = map[string][]string{
    // ... existing groups
    "NewGroup": {"alias1", "alias2", "alias3"},
}
```

### Step 2: Update Display Order

In `internal/application/service.go`:

```go
func knownGroupNames() []string {
    return []string{"Mono", "Optimized", "Low-Cost", "NewGroup", "Custom"}
}
```

### Step 3: Update Tests

Update all tests that check group counts or ordering.

## Config File Naming Convention

```
omo-<alias>.json
```

Where `<alias>` is:
- Lowercase
- Hyphens for multi-word names
- No spaces or special characters

Examples:
- `omo-claude.json` → alias: `claude`
- `omo-optimized-high.json` → alias: `optimized-high`
- `omo-lc-mode-low.json` → alias: `lc-mode-low`

## Active Config Detection

omo-switch determines the active config by:
1. Reading `~/.config/opencode/oh-my-openagent.json`
2. Comparing content with each discovered config
3. Returning the alias of the matching config

**Important**: Content comparison is byte-for-byte. Whitespace differences will cause mismatch.

## Backup System

Before each switch, omo-switch:
1. Creates backup of current active config
2. Stores in `~/.config/omo-switch/backups/`
3. Names as `oh-my-openagent.<timestamp>.json`

Backup is automatic and cannot be disabled.

## Integration Testing

### Test New Config

```bash
# Create test config
echo '{"agents":{"test":{}}}' > ~/.config/opencode/omo_configs/omo-test.json

# Verify discovery
./omo-switch --list

# Verify switching
./omo-switch test
./omo-switch --current

# Verify content
./omo-switch show test
```

### Test Validation

```bash
# Invalid config (missing agents)
echo '{"invalid":true}' > ~/.config/opencode/omo_configs/omo-bad.json

# Should show validation error
./omo-switch --list
```

## Common Integration Patterns

### Pattern 1: Model Family Configs

For configs with same provider but different models:

```json
{
  "omo-provider-fast.json": {
    "agents": { "sisyphus": { "model": "provider/fast-model" } }
  },
  "omo-provider-quality.json": {
    "agents": { "sisyphus": { "model": "provider/quality-model" } }
  }
}
```

### Pattern 2: Cost Tier Configs

For configs organized by cost:

```json
{
  "omo-budget.json": {
    "agents": { "sisyphus": { "model": "provider/cheap-model" } }
  },
  "omo-premium.json": {
    "agents": { "sisyphus": { "model": "provider/premium-model" } }
  }
}
```

### Pattern 3: Use Case Configs

For configs organized by use case:

```json
{
  "omo-coding.json": {
    "agents": { "sisyphus": { "model": "provider/code-model" } }
  },
  "omo-writing.json": {
    "agents": { "sisyphus": { "model": "provider/write-model" } }
  }
}
```

## Extending Validation

To add provider-specific validation:

### Option 1: Modify DefaultValidator

```go
func (v DefaultValidator) Validate(content []byte) error {
    // ... existing validation

    // Provider-specific validation
    agents, ok := parsed["agents"].(map[string]any)
    if ok {
        for name, agent := range agents {
            agentMap, ok := agent.(map[string]any)
            if !ok {
                continue
            }
            if _, ok := agentMap["model"]; !ok {
                return fmt.Errorf("agent %q missing model", name)
            }
        }
    }

    return nil
}
```

### Option 2: Create Custom Validator

```go
type ProviderValidator struct{}

func (v ProviderValidator) Validate(content []byte) error {
    // Custom validation logic
}

func (v ProviderValidator) RequiredKeys() []string {
    return []string{"agents"}
}
```

Then inject in main.go:

```go
validator := ProviderValidator{}
service := application.NewConfigService(store, backup, validator)
```

## Troubleshooting

### Config Not Appearing

1. Check filename matches `omo-*.json` pattern
2. Check file is in `~/.config/opencode/omo_configs/`
3. Check file is valid JSON
4. Run `./omo-switch --list` to see all discovered configs

### Validation Failing

1. Check JSON syntax
2. Check `agents` key exists
3. Run `./omo-switch show <alias>` to see content
4. Check for typos in key names

### Switching Not Working

1. Check config passes validation
2. Check file permissions
3. Check target file (`oh-my-openagent.json`) is writable
4. Check backup directory exists

## Reference

- [AGENTS.md](../AGENTS.md) - Master AI rules
- [cli-architecture.md](cli-architecture.md) - Architecture details
- [go-standards.md](go-standards.md) - Go coding standards

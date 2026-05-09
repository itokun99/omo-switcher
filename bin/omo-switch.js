#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const CONFIG_DIR = path.join(
  process.env.HOME,
  ".config/opencode/omo_configs"
);
const TARGET = path.join(
  process.env.HOME,
  ".config/opencode/oh-my-openagent.json"
);
const BACKUP_DIR = path.join(
  process.env.HOME,
  ".config/omo-switch/backups"
);

const KNOWN_GROUPS = {
  "Mono":      ["minimax", "qwen", "deepseek", "glm", "gpt", "claude"],
  "Optimized": ["optimized-high", "optimized-medium", "optimized-low"],
  "Low-Cost":  ["lc-mode-low", "lc-mode-medium", "lc-mode-high", "lc-mode-ultra"],
};

const REQUIRED_KEYS = ["agents"];

function validateSchema(filePath) {
  let raw;
  try {
    raw = fs.readFileSync(filePath, "utf8");
  } catch (e) {
    return { valid: false, error: `Cannot read file: ${e.message}` };
  }

  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (e) {
    return { valid: false, error: `Invalid JSON: ${e.message}` };
  }

  if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
    return { valid: false, error: "Config must be a JSON object" };
  }

  const missing = REQUIRED_KEYS.filter((k) => !(k in parsed));
  if (missing.length > 0) {
    return { valid: false, error: `Missing required keys: ${missing.join(", ")}` };
  }

  return { valid: true };
}

function backupCurrent() {
  if (!fs.existsSync(TARGET)) return null;

  fs.mkdirSync(BACKUP_DIR, { recursive: true });
  const ts = new Date().toISOString().replace(/[:.]/g, "-");
  const dest = path.join(BACKUP_DIR, `oh-my-openagent.${ts}.json`);
  fs.copyFileSync(TARGET, dest);
  return dest;
}

function discoverConfigs() {
  if (!fs.existsSync(CONFIG_DIR)) return {};
  return fs
    .readdirSync(CONFIG_DIR)
    .filter((f) => f.startsWith("omo-") && f.endsWith(".json"))
    .reduce((acc, file) => {
      const alias = file.replace(/^omo-/, "").replace(/\.json$/, "");
      acc[alias] = file;
      return acc;
    }, {});
}

function getCurrentAlias(configs) {
  if (!fs.existsSync(TARGET)) return null;
  const current = fs.readFileSync(TARGET, "utf8");
  for (const [alias, file] of Object.entries(configs)) {
    const src = path.join(CONFIG_DIR, file);
    if (fs.existsSync(src) && fs.readFileSync(src, "utf8") === current) {
      return alias;
    }
  }
  return null;
}

function listConfigs(configs) {
  const current = getCurrentAlias(configs);
  const knownAliases = new Set(Object.values(KNOWN_GROUPS).flat());
  const customAliases = Object.keys(configs).filter((a) => !knownAliases.has(a));

  console.log("\nAvailable configs:\n");

  for (const [group, aliases] of Object.entries(KNOWN_GROUPS)) {
    const present = aliases.filter((a) => configs[a]);
    if (present.length === 0) continue;
    console.log(`  ${group}`);
    for (const alias of present) {
      const marker = alias === current ? " ◀ active" : "";
      console.log(`    ${alias.padEnd(20)} → ${configs[alias]}${marker}`);
    }
  }

  if (customAliases.length > 0) {
    console.log(`  Custom`);
    for (const alias of customAliases.sort()) {
      const marker = alias === current ? " ◀ active" : "";
      console.log(`    ${alias.padEnd(20)} → ${configs[alias]}${marker}`);
    }
  }

  if (current === null && fs.existsSync(TARGET)) {
    console.log(`\n  Active config does not match any known file (manual edit?)`);
  }

  console.log();
}

function switchConfig(alias, configs) {
  if (!configs[alias]) {
    console.error(`\nUnknown config: "${alias}"\n`);
    listConfigs(configs);
    process.exit(1);
  }

  const src = path.join(CONFIG_DIR, configs[alias]);
  if (!fs.existsSync(src)) {
    console.error(`Source file not found: ${src}`);
    process.exit(1);
  }

  const check = validateSchema(src);
  if (!check.valid) {
    console.error(`\nSchema validation failed for "${alias}":`);
    console.error(`  ${check.error}\n`);
    process.exit(1);
  }

  const backupPath = backupCurrent();
  if (backupPath) {
    console.log(`Backed up current config → ${backupPath}`);
  }

  fs.copyFileSync(src, TARGET);
  console.log(`Switched to: ${alias} (${configs[alias]})`);
}

function showConfig(alias, configs) {
  if (!configs[alias]) {
    console.error(`Unknown config: "${alias}"`);
    process.exit(1);
  }
  const src = path.join(CONFIG_DIR, configs[alias]);
  if (!fs.existsSync(src)) {
    console.error(`File not found: ${src}`);
    process.exit(1);
  }
  const raw = fs.readFileSync(src, "utf8");
  console.log(`\n--- ${alias} (${configs[alias]}) ---\n`);
  console.log(raw);
}

function printHelp() {
  console.log(`
Usage: omo-switch [command] [alias]

Commands:
  (no args)          List all available configs
  --list, -l         List all available configs
  --current, -c      Show active config alias
  show <alias>       Print content of a config
  <alias>            Switch to that config

Examples:
  omo-switch                  # list all
  omo-switch claude           # switch to claude config
  omo-switch show optimized-high
  omo-switch --current
`);
}

const configs = discoverConfigs();
const args = process.argv.slice(2);

if (args.length === 0 || args[0] === "--list" || args[0] === "-l") {
  listConfigs(configs);
} else if (args[0] === "--current" || args[0] === "-c") {
  const cur = getCurrentAlias(configs);
  console.log(cur ? `Current: ${cur}` : "No active config matched");
} else if (args[0] === "--help" || args[0] === "-h") {
  printHelp();
} else if (args[0] === "show") {
  if (!args[1]) {
    console.error("Usage: omo-switch show <alias>");
    process.exit(1);
  }
  showConfig(args[1], configs);
} else {
  switchConfig(args[0], configs);
}

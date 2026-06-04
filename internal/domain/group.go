package domain

import "slices"

// Group represents a named collection of configs.
type Group struct {
	Name    string
	Configs []Config
}

// KnownGroups maps group names to their expected config aliases.
var KnownGroups = map[string][]string{
	"Mono":      {"minimax", "qwen", "deepseek", "glm", "gpt", "claude"},
	"Optimized": {"optimized-high", "optimized-medium", "optimized-low"},
	"Low-Cost":  {"lc-mode-low", "lc-mode-medium", "lc-mode-high", "lc-mode-ultra"},
}

// NewGroup creates an empty Group with the given name.
func NewGroup(name string) Group {
	return Group{Name: name}
}

// AddConfig appends a Config to the Group's Configs slice.
func (g *Group) AddConfig(c Config) {
	g.Configs = append(g.Configs, c)
}

// HasConfig reports whether a Config with the given alias exists in the Group.
func (g Group) HasConfig(alias string) bool {
	for _, c := range g.Configs {
		if c.Alias == alias {
			return true
		}
	}
	return false
}

// GetGroupForAlias returns the KnownGroups name that contains alias,
// or "Custom" if the alias is not in any known group.
func GetGroupForAlias(alias string) string {
	for group, aliases := range KnownGroups {
		if slices.Contains(aliases, alias) {
			return group
		}
	}
	return "Custom"
}

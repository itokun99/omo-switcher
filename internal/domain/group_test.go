package domain

import (
	"testing"
)

func TestNewGroup(t *testing.T) {
	g := NewGroup("TestGroup")

	if g.Name != "TestGroup" {
		t.Errorf("Name = %q, want %q", g.Name, "TestGroup")
	}
	if len(g.Configs) != 0 {
		t.Errorf("Configs length = %d, want 0", len(g.Configs))
	}
}

func TestGroup_AddConfig(t *testing.T) {
	g := NewGroup("Mono")
	c1 := NewConfig("claude", "omo-claude.json", "/tmp/omo-claude.json", nil)
	c2 := NewConfig("gpt", "omo-gpt.json", "/tmp/omo-gpt.json", nil)

	g.AddConfig(c1)
	if len(g.Configs) != 1 {
		t.Fatalf("after first add: Configs length = %d, want 1", len(g.Configs))
	}

	g.AddConfig(c2)
	if len(g.Configs) != 2 {
		t.Fatalf("after second add: Configs length = %d, want 2", len(g.Configs))
	}

	if g.Configs[0].Alias != "claude" {
		t.Errorf("Configs[0].Alias = %q, want %q", g.Configs[0].Alias, "claude")
	}
	if g.Configs[1].Alias != "gpt" {
		t.Errorf("Configs[1].Alias = %q, want %q", g.Configs[1].Alias, "gpt")
	}
}

func TestGroup_HasConfig(t *testing.T) {
	g := NewGroup("Test")
	g.AddConfig(NewConfig("claude", "omo-claude.json", "/tmp/omo-claude.json", nil))
	g.AddConfig(NewConfig("gpt", "omo-gpt.json", "/tmp/omo-gpt.json", nil))

	tests := []struct {
		name  string
		alias string
		want  bool
	}{
		{name: "existing alias", alias: "claude", want: true},
		{name: "another existing alias", alias: "gpt", want: true},
		{name: "non-existing alias", alias: "minimax", want: false},
		{name: "empty alias", alias: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.HasConfig(tt.alias); got != tt.want {
				t.Errorf("HasConfig(%q) = %v, want %v", tt.alias, got, tt.want)
			}
		})
	}
}

func TestGetGroupForAlias(t *testing.T) {
	tests := []struct {
		name  string
		alias string
		want  string
	}{
		{name: "mono alias - claude", alias: "claude", want: "Mono"},
		{name: "mono alias - minimax", alias: "minimax", want: "Mono"},
		{name: "mono alias - deepseek", alias: "deepseek", want: "Mono"},
		{name: "optimized alias - high", alias: "optimized-high", want: "Optimized"},
		{name: "optimized alias - medium", alias: "optimized-medium", want: "Optimized"},
		{name: "optimized alias - low", alias: "optimized-low", want: "Optimized"},
		{name: "low-cost alias - low", alias: "lc-mode-low", want: "Low-Cost"},
		{name: "low-cost alias - ultra", alias: "lc-mode-ultra", want: "Low-Cost"},
		{name: "custom alias", alias: "my-custom", want: "Custom"},
		{name: "empty alias", alias: "", want: "Custom"},
		{name: "unknown alias", alias: "nonexistent", want: "Custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGroupForAlias(tt.alias); got != tt.want {
				t.Errorf("GetGroupForAlias(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

// Package cli provides the CLI mode handler for backward compatibility
// with the Node.js omo-switch implementation.
package cli

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/itokun99/omo-switch/internal/application"
	"github.com/itokun99/omo-switch/internal/domain"
)

func Handle(service *application.ConfigService, args []string, w io.Writer) int {
	if len(args) == 0 {
		return cmdList(service, w)
	}

	switch args[0] {
	case "--list", "-l":
		return cmdList(service, w)
	case "--current", "-c":
		return cmdCurrent(service, w)
	case "--help", "-h":
		return cmdHelp(w)
	case "show":
		if len(args) < 2 {
			fmt.Fprintln(w, "Error: show requires an alias argument")
			return 1
		}
		return cmdShow(service, args[1], w)
	default:
		return cmdSwitch(service, args[0], w)
	}
}

func cmdList(service *application.ConfigService, w io.Writer) int {
	groups, err := service.ListConfigs()
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}

	current, err := service.GetActiveConfig()
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}

	fmt.Fprint(w, "\nAvailable configs:\n\n")

	for _, group := range groups {
		if len(group.Configs) == 0 {
			continue
		}

		fmt.Fprintf(w, "  %s\n", group.Name)

		sorted := sortGroupConfigs(group)
		for _, cfg := range sorted {
			marker := ""
			if cfg.Alias == current {
				marker = " \u25c0 active"
			}
			fmt.Fprintf(w, "    %-20s \u2192 %s%s\n", cfg.Alias, cfg.FileName, marker)
		}
	}

	return 0
}

func sortGroupConfigs(group domain.Group) []domain.Config {
	if group.Name == "Custom" {
		sorted := make([]domain.Config, len(group.Configs))
		copy(sorted, group.Configs)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Alias < sorted[j].Alias
		})
		return sorted
	}

	knownOrder, ok := domain.KnownGroups[group.Name]
	if !ok {
		return group.Configs
	}

	configMap := make(map[string]domain.Config, len(group.Configs))
	for _, cfg := range group.Configs {
		configMap[cfg.Alias] = cfg
	}

	var sorted []domain.Config
	for _, alias := range knownOrder {
		if cfg, ok := configMap[alias]; ok {
			sorted = append(sorted, cfg)
		}
	}

	for _, cfg := range group.Configs {
		if !slices.Contains(knownOrder, cfg.Alias) {
			sorted = append(sorted, cfg)
		}
	}

	return sorted
}

func cmdCurrent(service *application.ConfigService, w io.Writer) int {
	current, err := service.GetActiveConfig()
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}

	if current == "" {
		fmt.Fprintln(w, "No active config matched")
	} else {
		fmt.Fprintf(w, "Current: %s\n", current)
	}

	return 0
}

func cmdHelp(w io.Writer) int {
	fmt.Fprint(w, `
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
`)
	return 0
}

func cmdShow(service *application.ConfigService, alias string, w io.Writer) int {
	content, err := service.ShowConfig(alias)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}

	fileName := fmt.Sprintf("omo-%s.json", alias)
	fmt.Fprintf(w, "--- %s (%s) ---\n%s\n", alias, fileName, content)
	return 0
}

func cmdSwitch(service *application.ConfigService, alias string, w io.Writer) int {
	if strings.HasPrefix(alias, "-") {
		fmt.Fprintf(w, "Error: unknown option %s\n", alias)
		fmt.Fprintln(w, "Run 'omo-switch --help' for usage.")
		return 1
	}

	if err := service.SwitchConfig(alias); err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(w, "Switched to: %s\n", alias)
	return 0
}

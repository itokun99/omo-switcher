// Package components provides TUI sub-components for the omo-switch config switcher.
package components

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfigInfo holds metadata about a config file.
type ConfigInfo struct {
	Alias      string
	FileName   string
	FilePath   string
	FileSize   int64
	ModifiedAt string
	IsValid    bool
}

// InfoModel displays config metadata.
type InfoModel struct {
	info   ConfigInfo
	active bool
}

// NewInfoModel creates an InfoModel.
func NewInfoModel() InfoModel {
	return InfoModel{}
}

// Show activates the info view with the given config info.
func (m *InfoModel) Show(info ConfigInfo) {
	m.info = info
	m.active = true
}

// Hide deactivates the info view.
func (m *InfoModel) Hide() {
	m.active = false
}

// IsActive reports whether info view is active.
func (m InfoModel) IsActive() bool {
	return m.active
}

// GetConfigInfo retrieves metadata for a config file.
func GetConfigInfo(alias, filePath string) ConfigInfo {
	info := ConfigInfo{
		Alias:    alias,
		FilePath: filePath,
		FileName: fmt.Sprintf("omo-%s.json", alias),
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return info
	}

	info.FileSize = stat.Size()
	info.ModifiedAt = stat.ModTime().Format("2006-01-02 15:04:05")

	return info
}

// Render returns the info view string.
func (m InfoModel) Render(styles InfoStyles, width int) string {
	var b strings.Builder

	b.WriteString("  Config Info\n\n")

	fields := []struct {
		Label string
		Value string
	}{
		{"Alias", m.info.Alias},
		{"Filename", m.info.FileName},
		{"Path", m.info.FilePath},
		{"Size", formatSize(m.info.FileSize)},
		{"Modified", m.info.ModifiedAt},
	}

	for _, f := range fields {
		label := styles.Label.Render(fmt.Sprintf("  %-12s", f.Label))
		value := f.Value
		if value == "" {
			value = "(unknown)"
		}
		b.WriteString(label)
		b.WriteString(value)
		b.WriteString("\n")
	}

	return b.String()
}

// formatSize formats bytes as human-readable string.
func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
}

// InfoStyles holds styles for the info view.
type InfoStyles struct {
	Label lipgloss.Style
}

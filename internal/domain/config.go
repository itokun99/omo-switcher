// Package domain contains pure business logic types for the omo-switch TUI.
// No I/O, no filesystem access — only data structures and validation.
package domain

// Config represents a discovered omo config file.
type Config struct {
	Alias    string // e.g. "claude" from omo-claude.json
	FileName string // e.g. "omo-claude.json"
	FilePath string // full path: ~/.config/opencode/omo_configs/omo-claude.json
	Content  []byte // raw JSON bytes
	IsValid  bool   // schema validation result
	Error    string // validation error message if invalid
}

// NewConfig creates a Config with the given attributes.
// IsValid defaults to false; call Validate to check schema.
func NewConfig(alias, fileName, filePath string, content []byte) Config {
	return Config{
		Alias:    alias,
		FileName: fileName,
		FilePath: filePath,
		Content:  content,
	}
}

// Validate returns a new Config with IsValid and Error set based on
// the validator's assessment of Content. The original Config is not modified.
func (c Config) Validate(validator SchemaValidator) Config {
	err := validator.Validate(c.Content)
	if err != nil {
		return Config{
			Alias:    c.Alias,
			FileName: c.FileName,
			FilePath: c.FilePath,
			Content:  c.Content,
			IsValid:  false,
			Error:    err.Error(),
		}
	}
	return Config{
		Alias:    c.Alias,
		FileName: c.FileName,
		FilePath: c.FilePath,
		Content:  c.Content,
		IsValid:  true,
	}
}

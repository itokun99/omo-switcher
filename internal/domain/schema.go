package domain

import (
	"encoding/json"
	"fmt"
)

// SchemaValidator validates config content.
type SchemaValidator interface {
	Validate(content []byte) error
	RequiredKeys() []string
}

// DefaultValidator checks for required keys in JSON config.
type DefaultValidator struct{}

// var _ SchemaValidator = (*DefaultValidator)(nil) — compile-time check
var _ SchemaValidator = DefaultValidator{}

// Validate parses content as JSON and checks for the "agents" key.
// Returns an error if the JSON is invalid, not an object, or missing "agents".
func (v DefaultValidator) Validate(content []byte) error {
	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if _, ok := parsed["agents"]; !ok {
		return fmt.Errorf("missing required key: agents")
	}

	return nil
}

// RequiredKeys returns the list of keys that must be present in a valid config.
func (v DefaultValidator) RequiredKeys() []string {
	return []string{"agents"}
}

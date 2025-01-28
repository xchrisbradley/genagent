package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfigSpec defines a configuration requirement for a plugin
type ConfigSpec struct {
	Key          string   // Unique identifier for this config item
	Description  string   // Human-readable description
	Type         string   // Type of value (string, number, boolean)
	Required     bool     // Whether this config is required
	DefaultValue string   // Default value if not specified
	Options      []string // Possible values for enum types
}

// ConfigResponse represents user-provided configuration values
type ConfigResponse struct {
	Values map[string]string
}

// NewConfigResponse creates a new configuration response
func NewConfigResponse() *ConfigResponse {
	return &ConfigResponse{
		Values: make(map[string]string),
	}
}

// GetConfigFromUser prompts the user for configuration values based on specs
func GetConfigFromUser(specs []ConfigSpec) (*ConfigResponse, error) {
	reader := bufio.NewReader(os.Stdin)
	response := NewConfigResponse()

	for _, spec := range specs {
		// Skip if default value exists and not required
		if !spec.Required && spec.DefaultValue != "" {
			response.Values[spec.Key] = spec.DefaultValue
			continue
		}

		// Build prompt
		prompt := spec.Description
		if spec.DefaultValue != "" {
			prompt += fmt.Sprintf(" (default: %s)", spec.DefaultValue)
		}
		if len(spec.Options) > 0 {
			prompt += fmt.Sprintf(" [%s]", strings.Join(spec.Options, "/"))
		}
		prompt += ": "

		// Get user input
		fmt.Print(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading input: %v", err)
		}

		// Clean input
		input = strings.TrimSpace(input)

		// Use default if input is empty
		if input == "" && spec.DefaultValue != "" {
			input = spec.DefaultValue
		}

		// Validate required
		if spec.Required && input == "" {
			return nil, fmt.Errorf("required config '%s' not provided", spec.Key)
		}

		// Validate options
		if len(spec.Options) > 0 {
			valid := false
			for _, opt := range spec.Options {
				if input == opt {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("invalid value for '%s': must be one of %v", spec.Key, spec.Options)
			}
		}

		response.Values[spec.Key] = input
	}

	return response, nil
}

package policy

import (
	"encoding/json"

	"encore.app/pipeline/nodes"
)

// PolicyNode represents a node in the policy
type PolicyNode struct {
	ID     string          `json:"id"`
	Type   nodes.Type      `json:"type"`
	Config json.RawMessage `json:"config"`
	Next   []string        `json:"next"`
}

// PolicyDefinition represents a user-defined policy
type PolicyDefinition struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Nodes       map[string]PolicyNode `json:"nodes"`
	EntryPoints []string              `json:"entryPoints"`
}

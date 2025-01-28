package nodes

import (
	"encoding/json"

	"go.temporal.io/sdk/workflow"
)

// Type represents different types of pipeline nodes
type Type string

const (
	TypeHTTP Type = "http"
	// Add more node types as needed
)

// NodeConfig is a generic interface for node configurations
type Config interface {
	Validate() error
}

// NodeResult represents the result of a node execution
type NodeResult struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// Executor defines the interface for node execution
type Executor interface {
	// Execute runs the node with the given configuration
	Execute(ctx workflow.Context, config json.RawMessage) (*NodeResult, error)
	// ValidateConfig validates the node configuration
	ValidateConfig(config json.RawMessage) error
}

// Registry maintains a map of node types to their executors
type Registry struct {
	executors map[Type]Executor
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[Type]Executor),
	}
}

// Register adds a new node executor to the registry
func (r *Registry) Register(nodeType Type, executor Executor) {
	r.executors[nodeType] = executor
}

// Get returns the executor for the given node type
func (r *Registry) Get(nodeType Type) (Executor, bool) {
	executor, ok := r.executors[nodeType]
	return executor, ok
}

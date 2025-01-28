package pipeline

import (
	"encore.app/pipeline/nodes"
)

var nodeRegistry *nodes.Registry

// initRegistry initializes the node registry with all available node types
func initRegistry() *nodes.Registry {
	registry := nodes.NewRegistry()

	// Register HTTP node executor
	registry.Register(nodes.TypeHTTP, nodes.NewHTTPNodeExecutor())

	return registry
}

func init() {
	nodeRegistry = initRegistry()
}

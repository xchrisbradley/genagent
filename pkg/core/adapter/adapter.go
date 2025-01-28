package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"encore.app/pkg/core"
)

// RepoAdapter handles the conversion of an existing repository into a GenAgent-compatible project
type RepoAdapter struct {
	world       *core.World
	basePath    string
	config      *core.ConfigResponse
	pluginPaths []string
}

// NewRepoAdapter creates a new repository adapter
func NewRepoAdapter(basePath string) (*RepoAdapter, error) {
	world := core.NewWorld()
	return &RepoAdapter{
		world:    world,
		basePath: basePath,
	}, nil
}

// Initialize sets up the necessary GenAgent structure in the repository
func (ra *RepoAdapter) Initialize() error {
	// Create .genagent directory if it doesn't exist
	genagentDir := filepath.Join(ra.basePath, ".genagent")
	if err := os.MkdirAll(genagentDir, 0755); err != nil {
		return fmt.Errorf("failed to create .genagent directory: %v", err)
	}

	// Create plugins directory if it doesn't exist
	pluginsDir := filepath.Join(ra.basePath, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %v", err)
	}

	return nil
}

// AddPlugin registers a new plugin directory
func (ra *RepoAdapter) AddPlugin(pluginPath string) {
	ra.pluginPaths = append(ra.pluginPaths, pluginPath)
}

// Configure applies configuration to the adapter
func (ra *RepoAdapter) Configure(config *core.ConfigResponse) {
	ra.config = config
}

// Start initializes and runs all registered plugins
func (ra *RepoAdapter) Start() error {
	// Create a new entity for the repository
	// Initialize each plugin
	for _, pluginPath := range ra.pluginPaths {
		// Load and validate plugin
		_, err := core.LoadPlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin %s: %v", pluginPath, err)
		}

		// Create plugin component
		pluginComponent := &core.PluginComponent{
			Path: pluginPath,
		}
		// Set the type using the Type() method instead of direct field assignment
		pluginComponent.Type = "unknown" // This will be determined by plugin metadata
		// Cannot add component directly - PluginComponent needs to implement Component interface
		// TODO: Implement Type() method in PluginComponent to satisfy Component interface
		// For now, skip adding component
		// ra.world.AddComponent(entity, pluginComponent)
	}

	return nil
}

package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xchrisbradley/genagent/pkg/core"
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
	_ = ra.world.CreateEntity()

	// Initialize each plugin
	for _, pluginPath := range ra.pluginPaths {
		// Load and validate plugin
		pluginDir := filepath.Join(ra.basePath, "plugins", filepath.Base(pluginPath))

		// Create plugin directory if it doesn't exist
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory %s: %v", pluginDir, err)
		}

		// TODO: Load plugin configuration and metadata
		// For now, just associate the plugin with the entity
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

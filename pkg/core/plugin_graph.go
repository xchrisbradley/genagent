package core

import (
	"fmt"
	"sync"

	"encore.app/pkg/plugin"
)

// ValidatePlugin validates that a plugin meets all required interfaces
func ValidatePlugin(p plugin.Plugin) error {
	if p == nil {
		return fmt.Errorf("plugin is nil")
	}
	if p.ID() == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	if p.Name() == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	return nil
}

// PluginNode represents a node in the plugin dependency graph
type PluginNode struct {
	Plugin      plugin.Plugin
	DependsOn   map[string]*PluginNode
	Dependents  map[string]*PluginNode
	Initialized bool
	mu          sync.RWMutex
}

// PluginGraph manages plugin dependencies and initialization order
type PluginGraph struct {
	Nodes  map[string]*PluginNode
	world  *World
	mu     sync.RWMutex
	logger *Logger
}

// NewPluginGraph creates a new plugin dependency graph
func NewPluginGraph(world *World) *PluginGraph {
	return &PluginGraph{
		Nodes:  make(map[string]*PluginNode),
		world:  world,
		logger: world.Logger,
	}
}

// AddPlugin adds a plugin to the dependency graph
func (pg *PluginGraph) AddPlugin(plugin plugin.Plugin) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	// Validate plugin
	if err := ValidatePlugin(plugin); err != nil {
		return fmt.Errorf("invalid plugin: %v", err)
	}

	// Check for duplicate
	if _, exists := pg.Nodes[plugin.ID()]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.ID())
	}

	// Create new node
	node := &PluginNode{
		Plugin:     plugin,
		DependsOn:  make(map[string]*PluginNode),
		Dependents: make(map[string]*PluginNode),
	}

	// Add to graph
	pg.Nodes[plugin.ID()] = node
	pg.logger.Info("Added plugin to graph: %s v%s - %s",
		plugin.Name(),
		plugin.Version(),
		plugin.Metadata().Description)

	return nil
}

// AddDependency adds a dependency relationship between plugins
func (pg *PluginGraph) AddDependency(dependentID, dependsOnID string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	dependent, exists := pg.Nodes[dependentID]
	if !exists {
		return fmt.Errorf("dependent plugin %s not found", dependentID)
	}

	dependsOn, exists := pg.Nodes[dependsOnID]
	if !exists {
		return fmt.Errorf("dependency plugin %s not found", dependsOnID)
	}

	// Add bidirectional relationship
	dependent.DependsOn[dependsOnID] = dependsOn
	dependsOn.Dependents[dependentID] = dependent

	return nil
}

// InitializePlugins initializes all plugins in dependency order
func (pg *PluginGraph) InitializePlugins() error {
	// Get initialization order
	order, err := pg.getInitializationOrder()
	if err != nil {
		return err
	}

	// Initialize plugins in order
	for _, pluginID := range order {
		node := pg.Nodes[pluginID]
		if err := pg.initializePlugin(node); err != nil {
			return err
		}
	}

	return nil
}

// initializePlugin initializes a single plugin
func (pg *PluginGraph) initializePlugin(node *PluginNode) error {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.Initialized {
		return nil
	}

	// Create entity
	entity := NewEntity()

	// Get and apply configuration if needed
	specs := node.Plugin.ConfigSpecs()
	if len(specs) > 0 {
		// Convert plugin.ConfigSpec to core.ConfigSpec
		coreSpecs := make([]ConfigSpec, len(specs))
		for i, spec := range specs {
			coreSpecs[i] = ConfigSpec{
				Key:         spec.Name,
				Type:        "default",
				Required:    spec.Required,
				Description: spec.Description,
			}
		}

		coreConfig, err := GetConfigFromUser(coreSpecs)
		if err != nil {
			return fmt.Errorf("error getting plugin configuration: %v", err)
		}

		// Convert core.ConfigResponse to plugin.ConfigResponse
		values := make(map[string]interface{})
		for k, v := range coreConfig.Values {
			values[k] = v
		}
		pluginConfig := &plugin.ConfigResponse{
			Values: values,
		}

		if err := node.Plugin.Configure(pluginConfig); err != nil {
			return fmt.Errorf("error configuring plugin: %v", err)
		}
	}

	// Initialize plugin
	if err := node.Plugin.Initialize(pg.world, entity); err != nil {
		return fmt.Errorf("error initializing plugin: %v", err)
	}

	node.Initialized = true
	return nil
}

// getInitializationOrder returns plugins in dependency order
func (pg *PluginGraph) getInitializationOrder() ([]string, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	order := make([]string, 0)

	// Visit all nodes
	for id := range pg.Nodes {
		if !visited[id] {
			if err := pg.visit(id, visited, temp, &order); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

// visit performs a topological sort using DFS
func (pg *PluginGraph) visit(id string, visited, temp map[string]bool, order *[]string) error {
	if temp[id] {
		return fmt.Errorf("circular dependency detected")
	}
	if visited[id] {
		return nil
	}

	temp[id] = true

	// Visit dependencies first
	for depID := range pg.Nodes[id].DependsOn {
		if err := pg.visit(depID, visited, temp, order); err != nil {
			return err
		}
	}

	temp[id] = false
	visited[id] = true
	*order = append(*order, id)

	return nil
}

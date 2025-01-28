package core

import (
	"fmt"
	"path/filepath"
)

// Plugin represents a modular extension that can be added to an agent
type Plugin interface {
	// ID returns a unique identifier for this plugin
	ID() string

	// Name returns a human-readable name for this plugin
	Name() string

	// Version returns the semantic version of this plugin
	Version() string

	// Initialize is called when the plugin is first loaded
	Initialize(world *World, entity Entity) error

	// Components returns any components this plugin needs to register
	Components() []Component

	// Systems returns any systems this plugin needs to register
	Systems() []System

	// Metadata returns additional information about the plugin
	Metadata() PluginMetadata

	// ConfigSpecs returns the configuration specifications for this plugin
	ConfigSpecs() []ConfigSpec

	// Configure applies the configuration response to the plugin
	Configure(config *ConfigResponse) error
}

// PluginMetadata contains additional information about a plugin
type PluginMetadata struct {
	Description    string            // Human-readable description
	Author         string            // Plugin author/organization
	Website        string            // Plugin documentation/homepage
	Tags           []string          // Categorization tags
	Dependencies   []string          // Other plugin IDs this plugin depends on
	Configuration  map[string]string // Plugin-specific configuration
	Documentation  string            // Usage documentation
	ExampleConfigs []string          // Example configuration snippets
}

// PluginRegistry manages the lifecycle of plugins
type PluginRegistry struct {
	plugins map[string]Plugin
	logger  *Logger
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	logger, err := NewLogger(filepath.Join(".genagent", "logs"), true)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return &PluginRegistry{
		plugins: make(map[string]Plugin),
		logger:  logger,
	}
}

// Register adds a plugin to the registry
func (r *PluginRegistry) Register(plugin Plugin) error {
	id := plugin.ID()
	r.logger.Debug("Registering plugin: %s (version %s)", id, plugin.Version())

	if _, exists := r.plugins[id]; exists {
		r.logger.Error("Plugin with ID %s already registered", id)
		return fmt.Errorf("plugin with ID %s already registered", id)
	}

	// Validate plugin metadata
	metadata := plugin.Metadata()
	if metadata.Author == "" {
		r.logger.Error("Plugin %s missing required author information", id)
		return fmt.Errorf("plugin %s missing required author information", id)
	}

	r.plugins[id] = plugin
	r.logger.Info("Successfully registered plugin: %s", id)
	r.logger.Debug("Plugin details - Name: %s, Author: %s, Tags: %v",
		plugin.Name(), metadata.Author, metadata.Tags)

	return nil
}

// Get returns a plugin by ID
func (r *PluginRegistry) Get(id string) (Plugin, bool) {
	plugin, exists := r.plugins[id]
	return plugin, exists
}

// List returns all registered plugins
func (r *PluginRegistry) List() []Plugin {
	var list []Plugin
	for _, plugin := range r.plugins {
		list = append(list, plugin)
	}
	return list
}

// FindByTag returns all plugins with the specified tag
func (r *PluginRegistry) FindByTag(tag string) []Plugin {
	var matches []Plugin
	for _, plugin := range r.plugins {
		for _, t := range plugin.Metadata().Tags {
			if t == tag {
				matches = append(matches, plugin)
				break
			}
		}
	}
	return matches
}

// ValidatePlugin performs basic validation of a plugin
func ValidatePlugin(plugin Plugin) error {
	if plugin.ID() == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	if plugin.Name() == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if plugin.Version() == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}
	return nil
}

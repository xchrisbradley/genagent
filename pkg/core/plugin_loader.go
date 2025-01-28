package core

import (
	"fmt"
	"path/filepath"

	"encore.app/pkg/plugin"
)

// PluginFactory is a function type that creates a new plugin instance
type PluginFactory func() plugin.Plugin

// pluginFactories stores registered plugin factories
var pluginFactories = make(map[string]PluginFactory)

// RegisterPlugin registers a plugin factory function
func RegisterPlugin(name string, factory PluginFactory) {
	pluginFactories[name] = factory
}

// LoadPlugin loads a plugin from the specified directory path.
// It creates a new instance of the plugin using the registered factory function.
func LoadPlugin(pluginPath string) (plugin.Plugin, error) {
	// Get the plugin name from the directory name
	pluginName := filepath.Base(pluginPath)

	// Look up the factory function
	factory, exists := pluginFactories[pluginName]
	if !exists {

		// Check again after initialization
		factory, exists = pluginFactories[pluginName]
		if !exists {
			return nil, fmt.Errorf("unknown plugin: %s", pluginName)
		}
	}

	// Create a new instance using the factory
	plugin := factory()

	// Log plugin loading
	fmt.Printf("Loaded plugin: %s v%s - %s\n", plugin.Name(), plugin.Version(), plugin.Metadata().Description)

	return plugin, nil
}

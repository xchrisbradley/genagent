package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
)

// PluginManager handles the loading and management of plugins
type PluginManager struct {
	world     *World
	pluginDir string
	plugins   map[string]Plugin
}

// NewPluginManager creates a new plugin manager instance
func NewPluginManager(world *World, pluginDir string) *PluginManager {
	return &PluginManager{
		world:     world,
		pluginDir: pluginDir,
		plugins:   make(map[string]Plugin),
	}
}

// LoadPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return fmt.Errorf("error reading plugin directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pm.pluginDir, entry.Name(), "plugin.so")
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			// Try to compile the plugin
			if err := pm.compilePlugin(entry.Name()); err != nil {
				return fmt.Errorf("error compiling plugin %s: %v", entry.Name(), err)
			}
		}

		// Load the plugin
		if err := pm.loadPlugin(entry.Name()); err != nil {
			return fmt.Errorf("error loading plugin %s: %v", entry.Name(), err)
		}
	}

	return nil
}

// compilePlugin compiles the plugin source code into a shared object file
func (pm *PluginManager) compilePlugin(pluginName string) error {
	pluginDir := filepath.Join(pm.pluginDir, pluginName)
	pluginSrc := filepath.Join(pluginDir, "plugin.go")
	pluginOut := filepath.Join(pluginDir, "plugin.so")

	// Build the plugin
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginOut, pluginSrc)
	cmd.Dir = pluginDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// loadPlugin loads a compiled plugin and initializes it
func (pm *PluginManager) loadPlugin(pluginName string) error {
	pluginPath := filepath.Join(pm.pluginDir, pluginName, "plugin.so")

	// Open the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("error opening plugin: %v", err)
	}

	// Look up the New function
	sym, err := p.Lookup("New")
	if err != nil {
		return fmt.Errorf("plugin does not export 'New' function: %v", err)
	}

	// Assert that the symbol is a function that returns a Plugin
	newFunc, ok := sym.(func() Plugin)
	if !ok {
		return fmt.Errorf("plugin symbol is not a function that returns Plugin interface")
	}

	// Create a new plugin instance
	plug := newFunc()

	// Create a default entity for the plugin
	entity := NewEntity()

	// Initialize the plugin with the default entity
	if err := plug.Initialize(pm.world, entity); err != nil {
		return fmt.Errorf("error initializing plugin: %v", err)
	}

	// Store the plugin
	pm.plugins[pluginName] = plug

	return nil
}

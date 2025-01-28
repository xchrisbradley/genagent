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
	logger    *Logger
}

// NewPluginManager creates a new plugin manager instance
func NewPluginManager(world *World, pluginDir string) *PluginManager {
	return &PluginManager{
		world:     world,
		pluginDir: pluginDir,
		plugins:   make(map[string]Plugin),
		logger:    world.Logger,
	}
}

// LoadPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	// Ensure plugin directory exists
	if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
		return fmt.Errorf("error creating plugin directory: %v", err)
	}

	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return fmt.Errorf("error reading plugin directory: %v", err)
	}

	pm.logger.Info("Found %d potential plugins", len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if plugin.go exists
		pluginSrcPath := filepath.Join(pm.pluginDir, entry.Name(), "plugin.go")
		if _, err := os.Stat(pluginSrcPath); os.IsNotExist(err) {
			pm.logger.Debug("Skipping directory %s: no plugin.go found", entry.Name())
			continue
		}

		// Check if plugin needs compilation
		pluginPath := filepath.Join(pm.pluginDir, entry.Name(), "plugin.so")
		compile := false

		if soStat, err := os.Stat(pluginPath); os.IsNotExist(err) {
			compile = true
		} else if err == nil {
			// Check if source is newer than compiled version
			if srcStat, err := os.Stat(pluginSrcPath); err == nil {
				if srcStat.ModTime().After(soStat.ModTime()) {
					compile = true
				}
			}
		}

		if compile {
			pm.logger.Debug("Compiling plugin: %s", entry.Name())
			if err := pm.compilePlugin(entry.Name()); err != nil {
				pm.logger.Error("Error compiling plugin %s: %v", entry.Name(), err)
				continue
			}
		}

		// Load the plugin
		pm.logger.Info("Loading plugin: %s", entry.Name())
		if err := pm.loadPlugin(entry.Name()); err != nil {
			pm.logger.Error("Error loading plugin %s: %v", entry.Name(), err)
			continue
		}
	}

	pm.logger.Info("Successfully loaded %d plugins", len(pm.plugins))
	for _, p := range pm.plugins {
		meta := p.Metadata()
		pm.logger.Info("Plugin: %s v%s - %s", p.Name(), p.Version(), meta.Description)
		pm.logger.Debug("  Author: %s", meta.Author)
		pm.logger.Debug("  Website: %s", meta.Website)
		pm.logger.Debug("  Tags: %v", meta.Tags)
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

	// Initialize the plugin
	if err := plug.Initialize(pm.world, entity); err != nil {
		return fmt.Errorf("error initializing plugin: %v", err)
	}

	// Store the plugin
	pm.plugins[plug.ID()] = plug

	return nil
}

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
	registry  *PluginRegistry
}

// NewPluginManager creates a new plugin manager instance
func NewPluginManager(world *World, pluginDir string) *PluginManager {
	return &PluginManager{
		world:     world,
		pluginDir: pluginDir,
		plugins:   make(map[string]Plugin),
		logger:    world.Logger,
		registry:  NewPluginRegistry(),
	}
}

// LoadPlugins loads all plugins from both local directory and registry
func (pm *PluginManager) LoadPlugins() error {
	// Load local plugins first
	if err := pm.loadLocalPlugins(); err != nil {
		return fmt.Errorf("error loading local plugins: %v", err)
	}

	// Load registry plugins
	registryPlugins := pm.registry.List()
	for _, plug := range registryPlugins {
		// Skip if plugin is already loaded locally
		if _, exists := pm.plugins[plug.ID()]; exists {
			pm.logger.Debug("Skipping registry plugin %s: already loaded locally", plug.ID())
			continue
		}

		// Initialize the plugin
		entity := NewEntity()
		if err := plug.Initialize(pm.world, entity); err != nil {
			pm.logger.Error("Error initializing registry plugin %s: %v", plug.ID(), err)
			continue
		}

		// Store the plugin
		pm.plugins[plug.ID()] = plug
		pm.logger.Info("Loaded registry plugin: %s", plug.ID())
	}

	// Log summary
	pm.logger.Info("Successfully loaded %d plugins total", len(pm.plugins))
	for _, p := range pm.plugins {
		meta := p.Metadata()
		pm.logger.Info("Plugin: %s v%s - %s", p.Name(), p.Version(), meta.Description)
		pm.logger.Debug("  Author: %s", meta.Author)
		pm.logger.Debug("  Website: %s", meta.Website)
		pm.logger.Debug("  Tags: %v", meta.Tags)
	}

	return nil
}

// loadLocalPlugins loads plugins from the local plugin directory
func (pm *PluginManager) loadLocalPlugins() error {
	// Ensure plugin directory exists
	if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
		return fmt.Errorf("error creating plugin directory: %v", err)
	}

	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return fmt.Errorf("error reading plugin directory: %v", err)
	}

	var pluginDirs []string

	// Helper function to validate plugin directory
	validatePluginDir := func(dir string) bool {
		pluginSrcPath := filepath.Join(pm.pluginDir, dir, "plugin.go")
		if _, err := os.Stat(pluginSrcPath); err != nil {
			return false
		}

		// Check for required plugin files/structure
		files, err := os.ReadDir(filepath.Join(pm.pluginDir, dir))
		if err != nil {
			return false
		}

		hasGoMod := false
		for _, file := range files {
			if file.Name() == "go.mod" {
				hasGoMod = true
				break
			}
		}

		return hasGoMod
	}

	// Scan for plugins in root directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this is a direct plugin directory
		if validatePluginDir(entry.Name()) {
			pluginDirs = append(pluginDirs, entry.Name())
			continue
		}

		// Check subdirectories for plugins
		subEntries, err := os.ReadDir(filepath.Join(pm.pluginDir, entry.Name()))
		if err != nil {
			pm.logger.Debug("Error reading subdirectory %s: %v", entry.Name(), err)
			continue
		}

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}

			subDir := filepath.Join(entry.Name(), subEntry.Name())
			if validatePluginDir(subDir) {
				pluginDirs = append(pluginDirs, subDir)
			}
		}
	}

	pm.logger.Info("Found %d potential plugins", len(pluginDirs))

	for _, pluginDir := range pluginDirs {
		// Check if plugin needs compilation
		pluginPath := filepath.Join(pm.pluginDir, pluginDir, "plugin.so")
		pluginSrcPath := filepath.Join(pm.pluginDir, pluginDir, "plugin.go")
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
			pm.logger.Debug("Compiling plugin: %s", pluginDir)
			if err := pm.compilePlugin(pluginDir); err != nil {
				pm.logger.Error("Error compiling plugin %s: %v", pluginDir, err)
				continue
			}
		}

		// Load the plugin
		pm.logger.Info("Loading plugin: %s", pluginDir)
		if err := pm.loadPlugin(pluginDir); err != nil {
			pm.logger.Error("Error loading plugin %s: %v", pluginDir, err)
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

	// Get configuration specs and prompt for values
	specs := plug.ConfigSpecs()
	if len(specs) > 0 {
		config, err := GetConfigFromUser(specs)
		if err != nil {
			return fmt.Errorf("error getting plugin configuration: %v", err)
		}

		// Apply configuration
		if err := plug.Configure(config); err != nil {
			return fmt.Errorf("error configuring plugin: %v", err)
		}
	}

	// Initialize the plugin
	if err := plug.Initialize(pm.world, entity); err != nil {
		return fmt.Errorf("error initializing plugin: %v", err)
	}

	// Store the plugin
	pm.plugins[plug.ID()] = plug

	return nil
}

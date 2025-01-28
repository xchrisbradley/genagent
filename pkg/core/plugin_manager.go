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
	// First try to load from root plugins directory
	rootPluginDir := "plugins"
	if entries, err := os.ReadDir(rootPluginDir); err == nil {
		var pluginDirs []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if validatePluginDir(filepath.Join(rootPluginDir, entry.Name())) {
				pluginDirs = append(pluginDirs, filepath.Join(rootPluginDir, entry.Name()))
				pm.logger.Debug("Found plugin in root directory: %s", entry.Name())
			}
		}
		if len(pluginDirs) > 0 {
			pm.logger.Info("Found %d plugins in root directory", len(pluginDirs))
			return pm.loadPluginsFromDirs(pluginDirs)
		}
	}

	// Fall back to .genagent directory if no plugins found in root
	if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
		return fmt.Errorf("error creating plugin directory: %v", err)
	}

	pm.logger.Debug("Scanning .genagent plugin directory: %s", pm.pluginDir)
	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return fmt.Errorf("error reading plugin directory: %v", err)
	}

	var pluginDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if validatePluginDir(filepath.Join(pm.pluginDir, entry.Name())) {
			pluginDirs = append(pluginDirs, filepath.Join(pm.pluginDir, entry.Name()))
			pm.logger.Debug("Found plugin in .genagent directory: %s", entry.Name())
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

// initializeGoModule creates or updates the go.mod file for a plugin
// initializeGoModule is no longer needed as plugins are now internal packages
func (pm *PluginManager) initializeGoModule(pluginDir string) error {
	return nil
}

// validatePluginDir checks if a directory contains required plugin files
func validatePluginDir(dirName string) bool {
	// A valid plugin directory must contain plugin.go
	pluginFile := filepath.Join(dirName, "plugin.go")
	if _, err := os.Stat(pluginFile); err != nil {
		return false
	}
	return dirName != "" && dirName != "." && dirName != ".."
}

// loadPluginsFromDirs loads plugins from the provided list of directories
func (pm *PluginManager) loadPluginsFromDirs(pluginDirs []string) error {
	for _, pluginDir := range pluginDirs {
		// Check if plugin needs compilation
		pluginPath := filepath.Join(pluginDir, "plugin.so")
		pluginSrcPath := filepath.Join(pluginDir, "plugin.go")
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
			pm.logger.Debug("Compiling plugin: %s", filepath.Base(pluginDir))
			if err := pm.compilePlugin(filepath.Base(pluginDir)); err != nil {
				pm.logger.Error("Error compiling plugin %s: %v", filepath.Base(pluginDir), err)
				continue
			}
		}

		// Load the plugin
		pm.logger.Info("Loading plugin: %s", filepath.Base(pluginDir))
		if err := pm.loadPlugin(filepath.Base(pluginDir)); err != nil {
			pm.logger.Error("Error loading plugin %s: %v", filepath.Base(pluginDir), err)
			continue
		}
	}

	return nil
}

// compilePlugin compiles a plugin from source
func (pm *PluginManager) compilePlugin(pluginName string) error {
	pluginDir := filepath.Join("plugins", pluginName)

	// Initialize or update go.mod if needed
	if err := pm.initializeGoModule(pluginDir); err != nil {
		return fmt.Errorf("failed to initialize go module: %v", err)
	}

	// Set up build command with proper flags for plugin
	buildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", "plugin.so", "-tags", "plugin", "main.go")
	buildCmd.Dir = pluginDir

	// Execute build
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to compile plugin: %v\nOutput: %s", err, output)
	}

	return nil
}

// loadPlugin loads and initializes a plugin
func (pm *PluginManager) loadPlugin(pluginName string) error {
	// Load the plugin using Go's plugin package
	plugPath := filepath.Join("plugins", pluginName, "plugin.so")
	plug, err := plugin.Open(plugPath)
	if err != nil {
		return fmt.Errorf("error loading plugin %s: %v", pluginName, err)
	}

	// Look up the New symbol
	newSymbol, err := plug.Lookup("New")
	if err != nil {
		return fmt.Errorf("plugin %s does not export 'New' symbol: %v", pluginName, err)
	}

	// Assert that the symbol is of the correct type
	newFunc, ok := newSymbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("plugin %s: 'New' symbol has wrong type", pluginName)
	}

	// Create a new plugin instance
	plugin := newFunc()

	// Register the plugin with the registry
	if err := pm.registry.Register(plugin); err != nil {
		return fmt.Errorf("error registering plugin %s: %v", pluginName, err)
	}

	// Create a default entity for the plugin
	entity := NewEntity()

	// Get configuration specs and prompt for values
	specs := plugin.ConfigSpecs()
	if len(specs) > 0 {
		config, err := GetConfigFromUser(specs)
		if err != nil {
			return fmt.Errorf("error getting plugin configuration: %v", err)
		}

		// Apply configuration
		if err := plugin.Configure(config); err != nil {
			return fmt.Errorf("error configuring plugin: %v", err)
		}
	}

	// Initialize the plugin
	if err := plugin.Initialize(pm.world, entity); err != nil {
		return fmt.Errorf("error initializing plugin: %v", err)
	}

	// Store the plugin
	pm.plugins[plugin.ID()] = plugin

	return nil
}

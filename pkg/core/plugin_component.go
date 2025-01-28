package core

// PluginComponent represents a plugin's metadata and configuration state
type PluginComponent struct {
	// Path is the filesystem path to the plugin
	Path string

	// Type indicates the plugin's type (e.g., "go", "node", etc.)
	Type string

	// Metadata contains additional plugin information
	Metadata PluginMetadata

	// Enabled indicates if the plugin is currently active
	Enabled bool

	// Config stores plugin-specific configuration
	Config map[string]interface{}
}

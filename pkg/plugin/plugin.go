package plugin

// Plugin represents a modular extension that can be added to an agent
type Plugin interface {
	// ID returns a unique identifier for this plugin
	ID() string

	// Name returns a human-readable name for this plugin
	Name() string

	// Version returns the semantic version of this plugin
	Version() string

	// Initialize is called when the plugin is first loaded
	Initialize(world World, entity Entity) error

	// Components returns any components this plugin needs to register
	Components() []Component

	// Systems returns any systems this plugin needs to register
	Systems() []System

	// Metadata returns additional information about the plugin
	Metadata() PluginMetadata

	// ConfigSpecs returns the configuration specifications for the plugin
	ConfigSpecs() []ConfigSpec

	// Configure applies the provided configuration to the plugin
	Configure(config *ConfigResponse) error
}

// PluginMetadata contains additional information about a plugin
type PluginMetadata struct {
	Description string   // Human-readable description
	Author      string   // Plugin author/organization
	Website     string   // Plugin website URL
	Tags        []string // Categorization tags
}

// World represents the game world that contains all entities and systems
type World interface {
	CreateEntity() Entity
}

// Entity represents a game object that can have components attached
type Entity interface {
	ID() string
}

// Component represents a piece of functionality that can be attached to an entity
type Component interface {
	Type() string
}

// System represents a piece of logic that operates on entities with specific components
type System interface {
	Update(world World)
}

// ConfigSpec defines a configuration parameter for a plugin
type ConfigSpec struct {
	Name        string
	Description string
	Required    bool
	Default     interface{}
}

// ConfigResponse contains the configuration values for a plugin
type ConfigResponse struct {
	Values map[string]interface{}
}

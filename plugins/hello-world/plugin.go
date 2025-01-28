package helloworld

import (
	"fmt"

	"encore.app/pkg/plugin"
)

type HelloWorldPlugin struct{}

func (p *HelloWorldPlugin) ID() string {
	return "hello-world"
}

func (p *HelloWorldPlugin) Name() string {
	return "Hello World Plugin"
}

func (p *HelloWorldPlugin) Version() string {
	return "1.0.0"
}

func (p *HelloWorldPlugin) Initialize(world plugin.World, entity plugin.Entity) error {
	fmt.Println("Hello, World! This is your first GenAgent plugin!")
	return nil
}

func (p *HelloWorldPlugin) Components() []plugin.Component {
	return nil // No components needed for this simple plugin
}

func (p *HelloWorldPlugin) Systems() []plugin.System {
	return nil // No systems needed for this simple plugin
}

func (p *HelloWorldPlugin) Metadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Description: "A simple hello world plugin for GenAgent",
		Author:      "GenAgent User",
		Website:     "https://encore.app",
		Tags:        []string{"example", "hello-world"},
	}
}

// ConfigSpecs returns the configuration specifications for the plugin
func (p *HelloWorldPlugin) ConfigSpecs() []plugin.ConfigSpec {
	return []plugin.ConfigSpec{} // No configuration needed for this simple plugin
}

// Configure handles plugin configuration during initialization
func (p *HelloWorldPlugin) Configure(config *plugin.ConfigResponse) error {
	return nil // No configuration needed for this simple plugin
}

// New creates a new instance of the HelloWorld plugin
func New() plugin.Plugin {
	return &HelloWorldPlugin{}
}

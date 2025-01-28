package helloworld

import (
	"fmt"

	"github.com/xchrisbradley/genagent/pkg/core"
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

func (p *HelloWorldPlugin) Initialize(world *core.World, entity core.Entity) error {
	fmt.Println("Hello, World! This is your first GenAgent plugin!")
	return nil
}

func (p *HelloWorldPlugin) Components() []core.Component {
	return nil // No components needed for this simple plugin
}

func (p *HelloWorldPlugin) Systems() []core.System {
	return nil // No systems needed for this simple plugin
}

func (p *HelloWorldPlugin) Metadata() core.PluginMetadata {
	return core.PluginMetadata{
		Description: "A simple hello world plugin for GenAgent",
		Author:      "GenAgent User",
		Website:     "https://github.com/xchrisbradley/genagent",
		Tags:        []string{"example", "hello-world"},
	}
}

// ConfigSpecs returns the configuration specifications for the plugin
func (p *HelloWorldPlugin) ConfigSpecs() []core.ConfigSpec {
	return []core.ConfigSpec{} // No configuration needed for this simple plugin
}

// Configure handles plugin configuration during initialization
func (p *HelloWorldPlugin) Configure(config *core.ConfigResponse) error {
	return nil // No configuration needed for this simple plugin
}

// New creates a new instance of the HelloWorld plugin
func New() core.Plugin {
	return &HelloWorldPlugin{}
}

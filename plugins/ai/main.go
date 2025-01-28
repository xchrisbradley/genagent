package main

import (
	"github.com/xchrisbradley/genagent/pkg/core"
)

type AIPlugin struct{}

func (p *AIPlugin) ID() string {
	return "ai"
}

func (p *AIPlugin) Name() string {
	return "AI Plugin"
}

func (p *AIPlugin) Version() string {
	return "1.0.0"
}

func (p *AIPlugin) Initialize(world *core.World, entity core.Entity) error {
	return nil
}

func (p *AIPlugin) Components() []core.Component {
	return nil
}

func (p *AIPlugin) Systems() []core.System {
	return nil
}

func (p *AIPlugin) Metadata() core.PluginMetadata {
	return core.PluginMetadata{
		Description: "AI plugin for GenAgent",
		Author:      "GenAgent",
		Website:     "https://github.com/xchrisbradley/genagent",
		Tags:        []string{"ai", "plugin"},
	}
}

func (p *AIPlugin) ConfigSpecs() []core.ConfigSpec {
	return []core.ConfigSpec{}
}

func (p *AIPlugin) Configure(config *core.ConfigResponse) error {
	return nil
}

func New() core.Plugin {
	return &AIPlugin{}
}

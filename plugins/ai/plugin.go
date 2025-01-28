package ai

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xchrisbradley/genagent/pkg/core"
	"github.com/xchrisbradley/genagent/plugins/ai/llm"
)

// Component represents AI capabilities
type Component struct {
	Provider     llm.Provider
	Messages     []llm.Message
	LastMessage  string
	LastResponse string
	LastUpdate   time.Time
}

// System processes AI interactions
type System struct {
	lastProcessTime time.Time
}

func (s *System) Update(world *core.World, dt float64) {
	aiType := reflect.TypeOf(&Component{})

	for _, entity := range world.Entities() {
		comp := world.GetComponent(entity, aiType)
		if comp == nil {
			continue
		}

		ai := comp.(*Component)

		// Only process if we have a new message
		if ai.LastMessage != "" && ai.LastUpdate.After(s.lastProcessTime) {
			fmt.Printf("\n[AI] Processing message: %s\n", ai.LastMessage)

			// Add message to history
			ai.Messages = append(ai.Messages, llm.Message{
				Role:    "user",
				Content: ai.LastMessage,
			})

			// Get response from provider
			response, err := ai.Provider.Process(context.Background(), ai.Messages)
			if err != nil {
				fmt.Printf("[Error] %v\n", err)
				response = "Sorry, I encountered an error processing your message."
			}

			// Store response
			ai.LastResponse = response
			fmt.Printf("Agent: %s\n", response)

			// Add response to history
			ai.Messages = append(ai.Messages, llm.Message{
				Role:    "assistant",
				Content: response,
			})

			// Clear the input message
			ai.LastMessage = ""
		}
	}

	s.lastProcessTime = time.Now()
}

// Plugin implements the core.Plugin interface
type Plugin struct {
	provider llm.Provider
}

func NewPlugin(provider llm.Provider) *Plugin {
	return &Plugin{
		provider: provider,
	}
}

func (p *Plugin) ID() string {
	return "ai"
}

func (p *Plugin) Name() string {
	return fmt.Sprintf("AI Plugin (%s)", p.provider.Name())
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Initialize(world *core.World, entity core.Entity) error {
	comp := &Component{
		Provider: p.provider,
		Messages: make([]llm.Message, 0),
	}
	world.AddComponent(entity, comp)
	return nil
}

func (p *Plugin) Components() []core.Component {
	return []core.Component{
		&Component{},
	}
}

func (p *Plugin) Systems() []core.System {
	return []core.System{
		&System{},
	}
}

func (p *Plugin) Metadata() core.PluginMetadata {
	return core.PluginMetadata{
		Description:  "Provides AI capabilities through various LLM providers",
		Author:       "GenAgent Community",
		Website:      "https://github.com/xchrisbradley/genagent/plugins/ai",
		Tags:         []string{"ai", "llm", "chat"},
		Dependencies: []string{},
		Documentation: `
# AI Plugin

This plugin enables AI capabilities through various LLM providers. Currently supports:
- OpenAI (GPT-3.5/4)
- Google Gemini

## Usage Example

` + "```go" + `
// Create a provider
provider := llm.NewOpenAIProvider(apiKey, "")

// Create and register the plugin
plugin := ai.NewPlugin(provider)
world.RegisterPlugin(plugin)
` + "```" + `
`,
	}
}

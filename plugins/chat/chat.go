package chat

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/xchrisbradley/genagent/pkg/core"
)

// ChatComponent represents chatbot capabilities
type ChatComponent struct {
	LastMessage     string
	LastResponse    string
	LastInteraction time.Time
	Responses       map[string][]string
}

// ChatSystem handles chat interactions
type ChatSystem struct {
	lastProcessTime time.Time
}

func (s *ChatSystem) Update(world *core.World, dt float64) {
	chatType := reflect.TypeOf(&ChatComponent{})

	for _, entity := range world.Entities() {
		chat := world.GetComponent(entity, chatType)
		if chat == nil {
			continue
		}

		cc := chat.(*ChatComponent)

		// Only process if we have a new message
		if cc.LastMessage != "" && cc.LastInteraction.After(s.lastProcessTime) {
			fmt.Printf("\n[Chat] Processing message: %s\n", cc.LastMessage)

			// Find matching response
			response := s.findResponse(cc)

			if response != "" {
				cc.LastResponse = response
				fmt.Printf("Agent: %s\n", response)
			}

			// Clear the message after processing
			cc.LastMessage = ""
		}
	}

	s.lastProcessTime = time.Now()
}

func (s *ChatSystem) findResponse(cc *ChatComponent) string {
	input := strings.ToLower(cc.LastMessage)

	// Try to find exact matches first
	for keyword, responses := range cc.Responses {
		if strings.Contains(input, keyword) {
			return responses[rand.Intn(len(responses))]
		}
	}

	// If no exact match, use a default response
	defaultResponses := []string{
		"Interesting! Tell me more about that.",
		"I'm not sure I understand. Could you rephrase that?",
		"That's something to think about!",
		"I'm learning new things from our conversation!",
	}

	return defaultResponses[rand.Intn(len(defaultResponses))]
}

// Plugin implements the core.Plugin interface
type Plugin struct{}

func (p *Plugin) ID() string {
	return "chat"
}

func (p *Plugin) Name() string {
	return "Simple Chat Plugin"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Initialize(world *core.World, entity core.Entity) error {
	chat := &ChatComponent{
		LastInteraction: time.Now(),
		Responses: map[string][]string{
			"hello": {
				"Hi there! How can I help you?",
				"Hello! Nice to meet you!",
				"Greetings! What's on your mind?",
			},
			"how are you": {
				"I'm doing great, thanks for asking!",
				"I'm functioning perfectly!",
				"All systems operational and ready to help!",
			},
			"bye": {
				"Goodbye! Have a great day!",
				"See you later!",
				"Bye! Come back soon!",
			},
			"help": {
				"I can chat with you about various topics!",
				"Try saying hello, asking how I am, or just chat!",
				"I'm here to help and chat with you!",
			},
		},
	}
	world.AddComponent(entity, chat)
	return nil
}

func (p *Plugin) Components() []core.Component {
	return []core.Component{
		&ChatComponent{},
	}
}

func (p *Plugin) Systems() []core.System {
	return []core.System{
		&ChatSystem{},
	}
}

func (p *Plugin) Metadata() core.PluginMetadata {
	return core.PluginMetadata{
		Description:  "A simple chat plugin that enables basic conversation capabilities",
		Author:       "GenAgent Community",
		Website:      "https://github.com/xchrisbradley/genagent/plugins/chat",
		Tags:         []string{"chat", "communication", "basic"},
		Dependencies: []string{},
		Configuration: map[string]string{
			"response_delay": "0", // Instant responses by default
		},
		Documentation: `
# Chat Plugin

This plugin adds basic chat capabilities to your agent. It supports:
- Keyword-based responses
- Multiple response variations
- Default fallback responses

## Usage Example

` + "```go" + `
// Register the plugin
registry := core.NewPluginRegistry()
chatPlugin := &chat.Plugin{}
registry.Register(chatPlugin)

// Initialize with an agent
world := core.NewWorld()
agent := world.CreateEntity()
chatPlugin.Initialize(world, agent)

// Register components and systems
for _, component := range chatPlugin.Components() {
    world.RegisterComponent(reflect.TypeOf(component))
}
for _, system := range chatPlugin.Systems() {
    world.RegisterSystem(system)
}

// Get chat component and send a message
chatComponent := world.GetComponent(agent, reflect.TypeOf(&chat.ChatComponent{}))
chatComponent.(*chat.ChatComponent).LastMessage = "hello"
` + "```" + `

## Configuration Options

- response_delay: Delay in milliseconds before responding (default: "0")
  Example: {"response_delay": "1000"} // 1 second delay
`,
		ExampleConfigs: []string{
			`{"response_delay": "1000"}  // Add 1 second delay before responses`,
			`{"response_delay": "500"}   // Add 0.5 second delay before responses`,
		},
	}
}

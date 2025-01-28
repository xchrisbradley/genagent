package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/xchrisbradley/genagent/ai/llm"
	"github.com/xchrisbradley/genagent/ai/storage"
	"github.com/xchrisbradley/genagent/pkg/core"
)

// Component represents AI capabilities
type Component struct {
	Provider     llm.Provider
	Messages     []llm.Message
	LastMessage  string
	LastResponse string
	LastUpdate   time.Time
	Store        *storage.MessageStore
}

// GetContextStats returns statistics about the current conversation context
func (c *Component) GetContextStats() string {
	var stats strings.Builder
	var totalTokens int
	var userMsgs, assistantMsgs, systemMsgs int
	var oldestMsg time.Time
	var newestMsg time.Time

	if len(c.Messages) > 0 {
		oldestMsg = c.Messages[0].Timestamp
		newestMsg = c.Messages[0].Timestamp
	}

	for _, msg := range c.Messages {
		// Count tokens (rough estimate: 4 chars = 1 token)
		tokens := len(msg.Content) / 4
		totalTokens += tokens

		// Count message types
		switch msg.Role {
		case "user":
			userMsgs++
		case "assistant":
			assistantMsgs++
		case "system":
			systemMsgs++
		}

		// Track conversation timespan
		if msg.Timestamp.Before(oldestMsg) {
			oldestMsg = msg.Timestamp
		}
		if msg.Timestamp.After(newestMsg) {
			newestMsg = msg.Timestamp
		}
	}

	stats.WriteString("\n=== Context Statistics ===\n")
	stats.WriteString(fmt.Sprintf("Total Messages: %d\n", len(c.Messages)))
	stats.WriteString(fmt.Sprintf("- User Messages: %d\n", userMsgs))
	stats.WriteString(fmt.Sprintf("- Assistant Messages: %d\n", assistantMsgs))
	stats.WriteString(fmt.Sprintf("- System Messages: %d\n", systemMsgs))
	stats.WriteString(fmt.Sprintf("Estimated Tokens: %d\n", totalTokens))

	if len(c.Messages) > 0 {
		timespan := newestMsg.Sub(oldestMsg)
		stats.WriteString(fmt.Sprintf("Conversation Timespan: %s\n", formatDuration(timespan)))
		stats.WriteString(fmt.Sprintf("Average Tokens per Message: %.1f\n", float64(totalTokens)/float64(len(c.Messages))))
	}

	return stats.String()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%d days", days)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	}
	return fmt.Sprintf("%.1f seconds", d.Seconds())
}

// Message represents a chat message with additional metadata
type Message struct {
	llm.Message
	Timestamp time.Time
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

			// Add message to history with timestamp
			ai.Messages = append(ai.Messages, llm.Message{
				Role:      "user",
				Content:   ai.LastMessage,
				Timestamp: time.Now(),
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

			// Add response to history with timestamp
			ai.Messages = append(ai.Messages, llm.Message{
				Role:      "assistant",
				Content:   response,
				Timestamp: time.Now(),
			})

			// Save messages to storage
			if err := ai.Store.SaveMessages(ai.Messages); err != nil {
				fmt.Printf("[Error] Failed to save messages: %v\n", err)
			}

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
	// Create storage in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	storageDir := filepath.Join(homeDir, ".genagent", "chat_history")
	store, err := storage.NewMessageStore(storageDir)
	if err != nil {
		return fmt.Errorf("failed to create message store: %v", err)
	}

	// Load existing messages
	messages, err := store.LoadMessages()
	if err != nil {
		return fmt.Errorf("failed to load messages: %v", err)
	}

	// If no existing messages, add initial system message
	if len(messages) == 0 {
		messages = append(messages, llm.Message{
			Role:      "system",
			Content:   "You are a helpful AI assistant. Be concise in your responses.",
			Timestamp: time.Now(),
		})
	}

	comp := &Component{
		Provider: p.provider,
		Messages: messages,
		Store:    store,
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
		Description:  "Provides AI capabilities with infinite context window through various LLM providers",
		Author:       "GenAgent Community",
		Website:      "https://github.com/xchrisbradley/genagent/plugins/ai",
		Tags:         []string{"ai", "llm", "chat", "infinite-context"},
		Dependencies: []string{},
		Documentation: `
# AI Plugin with Infinite Context Window

This plugin enables AI capabilities through various LLM providers with support for infinite conversation history.

## Features
- Multiple LLM provider support (OpenAI, Gemini)
- Infinite context window through persistent storage
- Conversation statistics and analytics
- Message timestamps and metadata tracking
- Context-aware responses

## Usage Example

` + "```go" + `
// Create a provider
provider := llm.NewOpenAIProvider(apiKey, "")

// Create and register the plugin
plugin := ai.NewPlugin(provider)
world.RegisterPlugin(plugin)

// Get context statistics
aiComponent := world.GetComponent(entity, reflect.TypeOf(&ai.Component{})).(*ai.Component)
stats := aiComponent.GetContextStats()
fmt.Println(stats)
` + "```" + `
`,
	}
}

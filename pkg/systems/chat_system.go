package systems

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/xchrisbradley/genagent/pkg/components"
	"github.com/xchrisbradley/genagent/pkg/core"
)

// ChatSystem handles chat interactions
type ChatSystem struct {
	lastProcessTime time.Time
}

func (s *ChatSystem) Update(world *core.World, dt float64) {
	chatType := reflect.TypeOf(&components.ChatComponent{})
	thinkingType := reflect.TypeOf(&components.ThinkingComponent{})

	for _, entity := range world.Entities() {
		chat := world.GetComponent(entity, chatType)
		thinking := world.GetComponent(entity, thinkingType)

		if chat == nil || thinking == nil {
			continue
		}

		cc := chat.(*components.ChatComponent)
		tc := thinking.(*components.ThinkingComponent)

		// Only process if we have a new message
		if cc.LastMessage != "" && cc.LastInteraction.After(s.lastProcessTime) {
			fmt.Printf("\n[Agent %d] Processing message: %s\n", entity, cc.LastMessage)

			// Find matching response
			response := s.findResponse(cc, tc)

			if response != "" {
				cc.LastResponse = response
				fmt.Printf("[Agent %d] Response: %s\n", entity, response)

				// Update thinking context
				tc.Context["last_chat"] = map[string]interface{}{
					"message":  cc.LastMessage,
					"response": response,
					"time":     time.Now(),
				}
			}

			// Clear the message after processing
			cc.LastMessage = ""
		}
	}

	s.lastProcessTime = time.Now()
}

func (s *ChatSystem) findResponse(cc *components.ChatComponent, tc *components.ThinkingComponent) string {
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

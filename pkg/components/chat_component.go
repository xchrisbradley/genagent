package components

import "time"

// ChatComponent represents chatbot capabilities
type ChatComponent struct {
	LastMessage     string
	LastResponse    string
	LastInteraction time.Time
	Responses       map[string][]string // Map of keywords to possible responses
}

// NewChatComponent creates a new chat component with predefined responses
func NewChatComponent() *ChatComponent {
	return &ChatComponent{
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
			"weather": {
				"I can't check the actual weather, but I hope it's nice!",
				"Lovely weather for a chat, isn't it?",
				"Rain or shine, I'm here to talk!",
			},
			"name": {
				"You can call me GenAgent!",
				"I'm GenAgent, your friendly chat companion!",
				"GenAgent at your service!",
			},
		},
	}
}

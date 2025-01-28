package types

import "time"

// Bot represents a chat bot profile
type Bot struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Persona    string         `json:"persona"`
	Avatar     string         `json:"avatar,omitempty"`
	Provider   string         `json:"provider"`
	Parameters *BotParameters `json:"parameters,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// BotParameters represents configurable parameters for a bot
type BotParameters struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Platform  string    `json:"platform"` // discord, slack, local
	BotIDs    []string  `json:"bot_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	ChannelID      string    `json:"channel_id"`
	Platform       string    `json:"platform"`
	UserID         string    `json:"user_id"`
	BotID          string    `json:"bot_id,omitempty"`
	Content        string    `json:"content"`
	Type           string    `json:"type"` // text, image, etc
	CreatedAt      time.Time `json:"created_at"`
}

// ChatEvent represents a chat event for pub/sub
type ChatEvent struct {
	EventID   string    `json:"event_id"`
	Type      string    `json:"type"` // message, typing, etc
	Platform  string    `json:"platform"`
	ChannelID string    `json:"channel_id"`
	Message   *Message  `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

package types

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Parameters holds LLM generation parameters
type Parameters struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	Name() string
	GenerateResponse(ctx context.Context, messages []Message, params Parameters) (string, error)
}

// ProviderFactory creates Provider instances
type ProviderFactory interface {
	Create(apiKey string) (Provider, error)
}

// LLMRequestEvent represents an LLM request
type LLMRequestEvent struct {
	RequestID      string     `json:"request_id"`
	BotID          string     `json:"bot_id"`
	ChannelID      string     `json:"channel_id"`
	ConversationID string     `json:"conversation_id"`
	Provider       string     `json:"provider"`
	Messages       []Message  `json:"messages"`
	Parameters     Parameters `json:"parameters"`
	Timestamp      time.Time
}

func (e *LLMRequestEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// LLMResponseEvent represents an LLM response
type LLMResponseEvent struct {
	RequestID      string `json:"request_id"`
	BotID          string `json:"bot_id"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content,omitempty"`
	Error          string `json:"error,omitempty"`
	Timestamp      time.Time
}

func (e *LLMResponseEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// NewLLMResponseEvent creates a new LLMResponseEvent
func NewLLMResponseEvent(requestID, botID, conversationID, content string, err error) *LLMResponseEvent {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	return &LLMResponseEvent{
		RequestID:      requestID,
		BotID:          botID,
		ConversationID: conversationID,
		Content:        content,
		Error:          errStr,
		Timestamp:      time.Now(),
	}
}

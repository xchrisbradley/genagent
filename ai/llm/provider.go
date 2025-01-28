package llm

import (
	"context"
	"fmt"
	"time"
)

// Message represents a chat message
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	// Process takes a list of messages and returns a response
	Process(ctx context.Context, messages []Message) (string, error)
	// Name returns the provider's name
	Name() string
}

// BaseProvider provides common functionality for providers
type BaseProvider struct {
	name      string
	maxTokens int
}

func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		name:      name,
		maxTokens: 150,
	}
}

func (p *BaseProvider) Name() string {
	return p.name
}

func (p *BaseProvider) SetMaxTokens(tokens int) {
	p.maxTokens = tokens
}

// ProviderError represents an error from an LLM provider
type ProviderError struct {
	Provider string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s provider error: %s (%v)", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("%s provider error: %s", e.Provider, e.Message)
}

func NewProviderError(provider string, message string, err error) error {
	return &ProviderError{
		Provider: provider,
		Message:  message,
		Err:      err,
	}
}

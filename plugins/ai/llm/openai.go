package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	*BaseProvider
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, model string) *OpenAIProvider {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	return &OpenAIProvider{
		BaseProvider: NewBaseProvider(fmt.Sprintf("OpenAI (%s)", model)),
		client:       openai.NewClient(apiKey),
		model:        model,
	}
}

func (p *OpenAIProvider) Process(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     p.model,
			Messages:  openaiMessages,
			MaxTokens: p.maxTokens,
		},
	)

	if err != nil {
		return "", NewProviderError("OpenAI", "API request failed", err)
	}

	if len(resp.Choices) == 0 {
		return "", NewProviderError("OpenAI", "no completion choices returned", nil)
	}

	return resp.Choices[0].Message.Content, nil
}

// SetModel changes the OpenAI model
func (p *OpenAIProvider) SetModel(model string) {
	p.model = model
	p.name = fmt.Sprintf("OpenAI (%s)", model)
}

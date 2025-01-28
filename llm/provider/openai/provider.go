package openai

import (
	"context"
	"fmt"

	"encore.app/llm/types"
	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *openai.Client
}

type Factory struct{}

func (f *Factory) Create(apiKey string) (types.Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(apiKey)
	return &Provider{
		client: client,
	}, nil
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) GenerateResponse(ctx context.Context, messages []types.Message, params types.Parameters) (string, error) {
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
			Model:       openai.GPT3Dot5Turbo,
			Messages:    openaiMessages,
			MaxTokens:   params.MaxTokens,
			Temperature: float32(params.Temperature),
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices")
	}

	return resp.Choices[0].Message.Content, nil
}

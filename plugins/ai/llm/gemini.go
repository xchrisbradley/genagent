package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements Provider interface using Google's Gemini
type GeminiProvider struct {
	*BaseProvider
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, NewProviderError("Gemini", "failed to create client", err)
	}

	return &GeminiProvider{
		BaseProvider: NewBaseProvider("Google Gemini"),
		client:       client,
		model:        client.GenerativeModel("gemini-pro"),
	}, nil
}

func (p *GeminiProvider) Process(ctx context.Context, messages []Message) (string, error) {
	// Build conversation history
	var prompt string
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt += fmt.Sprintf("System: %s\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("User: %s\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("Assistant: %s\n", msg.Content)
		}
	}

	resp, err := p.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", NewProviderError("Gemini", "generation failed", err)
	}

	if len(resp.Candidates) == 0 {
		return "", NewProviderError("Gemini", "no response generated", nil)
	}

	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", NewProviderError("Gemini", "empty response", nil)
	}

	var responses []string
	for _, part := range candidate.Content.Parts {
		switch p := part.(type) {
		case *genai.Text:
			responses = append(responses, string(*p))
		}
	}

	if len(responses) == 0 {
		return "", NewProviderError("Gemini", "no text in response", nil)
	}

	return strings.Join(responses, " "), nil
}

// Close cleans up the Gemini client
func (p *GeminiProvider) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

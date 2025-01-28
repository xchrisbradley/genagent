package gemini

import (
	"context"
	"fmt"

	"encore.app/llm/types"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Provider struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

type Factory struct{}

func (f *Factory) Create(apiKey string) (types.Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-pro")
	return &Provider{
		client: client,
		model:  model,
	}, nil
}

func (p *Provider) Name() string {
	return "gemini"
}

func (p *Provider) GenerateResponse(ctx context.Context, messages []types.Message, params types.Parameters) (string, error) {
	// Convert messages to Gemini format
	var prompt []genai.Part
	for _, msg := range messages {
		prompt = append(prompt, genai.Text(msg.Content))
	}

	// Configure generation parameters
	p.model.SetTemperature(float32(params.Temperature))
	if params.MaxTokens > 0 {
		p.model.SetMaxOutputTokens(int32(params.MaxTokens))
	}

	// Generate response
	resp, err := p.model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}

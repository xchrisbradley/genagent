package togetherai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"encore.app/llm/types"
)

type Provider struct {
	apiKey string
	client *http.Client
	model  string
}

type Factory struct{}

func (f *Factory) Create(apiKey string) (types.Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("TogetherAI API key is required")
	}

	return &Provider{
		apiKey: apiKey,
		client: &http.Client{},
		model:  "meta-llama/Llama-3.3-70B-Instruct-Turbo-Free", // Default model
	}, nil
}

func (p *Provider) Name() string {
	return "togetherai"
}

func (p *Provider) GenerateResponse(ctx context.Context, messages []types.Message, params types.Parameters) (string, error) {
	// Convert messages to TogetherAI format
	var prompt string
	for _, msg := range messages {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":       p.model,
		"prompt":      prompt,
		"temperature": params.Temperature,
		"max_tokens":  params.MaxTokens,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.together.xyz/inference", strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Output struct {
			Choices []struct {
				Text string `json:"text"`
			} `json:"choices"`
		} `json:"output"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Output.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return result.Output.Choices[0].Text, nil
}

package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	*BaseProvider
	apiKey     string
	model      string
	httpClient *http.Client
}

// OpenAI API request/response types
type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, model string) Provider {
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &OpenAIProvider{
		BaseProvider: NewBaseProvider("OpenAI"),
		apiKey:       apiKey,
		model:        model,
		httpClient:   &http.Client{},
	}
}

// Process implements the Provider interface
func (p *OpenAIProvider) Process(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to OpenAI format
	openAIMessages := make([]openAIMessage, len(messages))
	for i, msg := range messages {
		openAIMessages[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	reqBody := openAIRequest{
		Model:     p.model,
		Messages:  openAIMessages,
		MaxTokens: p.maxTokens,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError("OpenAI", "Failed to marshal request", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.openai.com/v1/chat/completions",
		strings.NewReader(string(reqData)),
	)
	if err != nil {
		return "", NewProviderError("OpenAI", "Failed to create request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", NewProviderError("OpenAI", "Failed to send request", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewProviderError("OpenAI", "Failed to read response", err)
	}

	// Parse response
	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", NewProviderError("OpenAI", "Failed to parse response", err)
	}

	// Check for API error
	if openAIResp.Error != nil {
		return "", NewProviderError("OpenAI", openAIResp.Error.Message, nil)
	}

	// Check for empty response
	if len(openAIResp.Choices) == 0 {
		return "", NewProviderError("OpenAI", "No response choices returned", nil)
	}

	return openAIResp.Choices[0].Message.Content, nil
}

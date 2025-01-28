package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// GeminiProvider implements the Provider interface for Google's Gemini
type GeminiProvider struct {
	*BaseProvider
	apiKey     string
	httpClient *http.Client
}

// Gemini API request/response types
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, apiKey string) (Provider, error) {
	return &GeminiProvider{
		BaseProvider: NewBaseProvider("Gemini"),
		apiKey:       apiKey,
		httpClient:   &http.Client{},
	}, nil
}

// Close implements io.Closer
func (p *GeminiProvider) Close() error {
	return nil
}

// Process implements the Provider interface
func (p *GeminiProvider) Process(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to Gemini format
	var contents []geminiContent
	for _, msg := range messages {
		contents = append(contents, geminiContent{
			Parts: []geminiPart{{Text: msg.Content}},
			Role:  msg.Role,
		})
	}

	// Create request
	reqBody := geminiRequest{
		Contents: contents,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError("Gemini", "Failed to marshal request", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key="+p.apiKey,
		strings.NewReader(string(reqData)),
	)
	if err != nil {
		return "", NewProviderError("Gemini", "Failed to create request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", NewProviderError("Gemini", "Failed to send request", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewProviderError("Gemini", "Failed to read response", err)
	}

	// Parse response
	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", NewProviderError("Gemini", "Failed to parse response", err)
	}

	// Check for API error
	if geminiResp.Error != nil {
		return "", NewProviderError("Gemini", geminiResp.Error.Message, nil)
	}

	// Check for empty response
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", NewProviderError("Gemini", "No response content returned", nil)
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

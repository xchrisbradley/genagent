package llm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"encore.app/llm/types"
)

type Action struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response,omitempty"`
	Error     string    `json:"error,omitempty"`
	Status    string    `json:"status"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateRequest represents the request parameters for generating a response
type GenerateRequest struct {
	Persona string `json:"persona"`
	Prompt  string `json:"prompt"`
}

//encore:api public method=POST path=/api/llm/generate
func (s *Service) GenerateResponse(ctx context.Context, params *GenerateRequest) (*Action, error) {
	// Generate request ID
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())

	// Create messages array with system and user messages
	messages := []types.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are %s. Respond in character.", params.Persona),
		},
		{
			Role:    "user",
			Content: params.Prompt,
		},
	}

	// Set default parameters from config
	llmParams := types.Parameters{
		MaxTokens:   int(cfg.MaxTokens()),
		Temperature: cfg.Temperature(),
	}

	// Create request event
	req := &types.LLMRequestEvent{
		RequestID:  requestID,
		BotID:      "api",
		ChannelID:  "api",
		Provider:   cfg.DefaultProvider(),
		Messages:   messages,
		Parameters: llmParams,
		Timestamp:  time.Now(),
	}

	// Process request through pub/sub
	if err := s.ProcessRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("process request: %w", err)
	}

	// Create and return initial action
	action := &Action{
		ID:        requestID,
		Type:      "LLMResponse",
		Prompt:    params.Prompt,
		Status:    "pending",
		Provider:  req.Provider,
		CreatedAt: time.Now(),
	}

	return action, nil
}

// GetGenerationStatus retrieves the status of a generation request
//
//encore:api public method=GET path=/api/llm/status/:id
func (s *Service) GetGenerationStatus(ctx context.Context, id string) (*Action, error) {
	var (
		messages, parameters []byte
		response             sql.NullString
		errorMsg             sql.NullString
		completedAt          sql.NullTime
	)

	query := `
		SELECT messages, parameters, response, error, completed_at
		FROM llm_requests
		WHERE request_id = $1
	`
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&messages, &parameters, &response, &errorMsg, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("generation request not found")
	} else if err != nil {
		return nil, fmt.Errorf("query generation status: %w", err)
	}

	// Parse messages to get the prompt
	var msgs []types.Message
	if err := json.Unmarshal(messages, &msgs); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}

	// Find user message for prompt
	var prompt string
	for _, msg := range msgs {
		if msg.Role == "user" {
			prompt = msg.Content
			break
		}
	}

	// Determine status
	status := "pending"
	if completedAt.Valid {
		status = "completed"
		if errorMsg.Valid {
			status = "failed"
		}
	}

	action := &Action{
		ID:        id,
		Type:      "LLMResponse",
		Prompt:    prompt,
		Status:    status,
		CreatedAt: time.Now(),
	}

	if response.Valid {
		action.Response = response.String
	}
	if errorMsg.Valid {
		action.Error = errorMsg.String
	}

	return action, nil
}

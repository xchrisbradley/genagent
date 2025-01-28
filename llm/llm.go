package llm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"encore.dev/config"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"

	"encore.app/llm/provider/gemini"
	"encore.app/llm/provider/openai"
	"encore.app/llm/provider/togetherai"
	llmpubsub "encore.app/llm/pubsub"
	"encore.app/llm/types"
)

// Define secrets and database
var secrets struct {
	OpenAIKey     string
	GeminiKey     string
	TogetherAIKey string
}

// Initialize database connection
var db = sqldb.NewDatabase("llm", sqldb.DatabaseConfig{
	Migrations: "db/migrations",
})

// LLMConfig defines configuration for the LLM service
type LLMConfig struct {
	DefaultProvider config.String
	MaxTokens       config.Int
	Temperature     config.Float64
}

// Load configuration
var cfg = config.Load[*LLMConfig]()

//encore:service
type Service struct {
	DB        *sql.DB
	Providers map[string]types.Provider
}

// Initialize service
func initService() (*Service, error) {
	s := &Service{
		DB:        db.Stdlib(),
		Providers: make(map[string]types.Provider),
	}

	// Initialize providers
	if secrets.OpenAIKey != "" {
		if p, err := (&openai.Factory{}).Create(secrets.OpenAIKey); err == nil {
			s.Providers[p.Name()] = p
		}
	}
	if secrets.GeminiKey != "" {
		if p, err := (&gemini.Factory{}).Create(secrets.GeminiKey); err == nil {
			s.Providers[p.Name()] = p
		}
	}
	if secrets.TogetherAIKey != "" {
		if p, err := (&togetherai.Factory{}).Create(secrets.TogetherAIKey); err == nil {
			s.Providers[p.Name()] = p
		}
	}

	return s, nil
}

// processGeneration handles an LLM generation request
func (s *Service) processGeneration(ctx context.Context, req *types.LLMRequestEvent) error {
	// Get provider
	p, ok := s.Providers[req.Provider]
	if !ok {
		// Publish error response
		respEvent := types.NewLLMResponseEvent(req.RequestID, req.BotID, req.ConversationID, "", fmt.Errorf("unknown provider: %s", req.Provider))
		_, pubErr := llmpubsub.GenerationResponses.Publish(ctx, respEvent)
		if pubErr != nil {
			return fmt.Errorf("publish error response: %w", pubErr)
		}
		return s.storeResponse(ctx, respEvent)
	}

	// Generate response
	response, err := p.GenerateResponse(ctx, req.Messages, req.Parameters)
	if err != nil {
		// Publish error response
		respEvent := types.NewLLMResponseEvent(req.RequestID, req.BotID, req.ConversationID, "", err)
		_, pubErr := llmpubsub.GenerationResponses.Publish(ctx, respEvent)
		if pubErr != nil {
			return fmt.Errorf("publish error response: %w", pubErr)
		}
		return s.storeResponse(ctx, respEvent)
	}

	// Publish success response
	respEvent := types.NewLLMResponseEvent(req.RequestID, req.BotID, req.ConversationID, response, nil)
	_, err = llmpubsub.GenerationResponses.Publish(ctx, respEvent)
	if err != nil {
		return fmt.Errorf("publish success response: %w", err)
	}
	return s.storeResponse(ctx, respEvent)
}

// storeResponse stores the response in the database
func (s *Service) storeResponse(ctx context.Context, resp *types.LLMResponseEvent) error {
	query := `
		UPDATE llm_requests 
		SET response = $1, error = $2, completed_at = $3
		WHERE request_id = $4
	`
	_, err := s.DB.ExecContext(ctx, query,
		resp.Content,
		sql.NullString{String: resp.Error, Valid: resp.Error != ""},
		resp.OccurredAt(),
		resp.RequestID,
	)
	return err
}

// Initialize subscriptions
var _ = pubsub.NewSubscription(
	llmpubsub.GenerationRequests, "process-generation",
	pubsub.SubscriptionConfig[*types.LLMRequestEvent]{
		Handler: pubsub.MethodHandler((*Service).processGeneration),
	},
)

var _ = pubsub.NewSubscription(
	llmpubsub.GenerationResponses, "store-response",
	pubsub.SubscriptionConfig[*types.LLMResponseEvent]{
		Handler: pubsub.MethodHandler((*Service).storeResponse),
	},
)

// GenerateProviderResponse generates a response using the specified provider
func (s *Service) GenerateProviderResponse(ctx context.Context, messages []types.Message, params types.Parameters) (string, error) {
	// Use default provider from config if available, otherwise use first available provider
	provider := cfg.DefaultProvider()
	if provider == "" {
		for name := range s.Providers {
			provider = name
			break
		}
	}

	p, ok := s.Providers[provider]
	if !ok {
		return "", fmt.Errorf("no provider available")
	}

	return p.GenerateResponse(ctx, messages, params)
}

// ProcessRequest publishes an LLM request to the generation topic
//
//encore:api public method=POST path=/api/llm/process
func (s *Service) ProcessRequest(ctx context.Context, req *types.LLMRequestEvent) error {
	// Store request in database first
	messagesJSON, err := json.Marshal(req.Messages)
	if err != nil {
		return fmt.Errorf("marshal messages: %w", err)
	}

	paramsJSON, err := json.Marshal(req.Parameters)
	if err != nil {
		return fmt.Errorf("marshal parameters: %w", err)
	}

	query := `
		INSERT INTO llm_requests (
			request_id, bot_id, channel_id, conversation_id,
			provider, messages, parameters, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.DB.ExecContext(ctx, query,
		req.RequestID,
		req.BotID,
		req.ChannelID,
		req.ConversationID,
		req.Provider,
		messagesJSON,
		paramsJSON,
		req.OccurredAt(),
	)
	if err != nil {
		return fmt.Errorf("store request: %w", err)
	}

	// Publish request to topic
	_, err = llmpubsub.GenerationRequests.Publish(ctx, req)
	if err != nil {
		return fmt.Errorf("publish request: %w", err)
	}

	return nil
}

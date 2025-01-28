package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"github.com/google/uuid"
	"github.com/lib/pq"

	chatpubsub "encore.app/chat/pubsub"
	"encore.app/chat/types"
	"encore.app/llm"
	llmpubsub "encore.app/llm/pubsub"
	llmtypes "encore.app/llm/types"
)

// Initialize database connection
var db = sqldb.NewDatabase("chat", sqldb.DatabaseConfig{
	Migrations: "db/migrations",
})

// BroadcastRequest represents the request for broadcasting a chat event
type BroadcastRequest struct {
	Event types.ChatEvent `json:"event"`
}

// Broadcast sends a message to all connected WebSocket clients
//
//encore:api public method=POST path=/api/chat/broadcast
func (s *Service) Broadcast(ctx context.Context, req *BroadcastRequest) error {
	event := req.Event

	// Set event metadata if not provided
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%s", uuid.New().String())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Publish to chat events topic
	_, err := chatpubsub.ChatEvents.Publish(ctx, &event)
	if err != nil {
		return fmt.Errorf("publish broadcast event: %w", err)
	}

	return nil
}

//encore:service
type Service struct {
	DB *sql.DB
}

// Initialize service
func initService() (*Service, error) {
	return &Service{
		DB: db.Stdlib(),
	}, nil
}

// CreateBot creates a new bot profile
//
//encore:api public method=POST path=/api/chat/bots
func (s *Service) CreateBot(ctx context.Context, bot *types.Bot) (*types.Bot, error) {
	if bot.ID == "" {
		bot.ID = fmt.Sprintf("bot_%s", uuid.New().String())
	}

	paramsJSON, err := json.Marshal(bot.Parameters)
	if err != nil {
		return nil, fmt.Errorf("marshal parameters: %w", err)
	}

	now := time.Now()
	bot.CreatedAt = now
	bot.UpdatedAt = now

	query := `
		INSERT INTO bots (
			id, name, persona, avatar, provider,
			parameters, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, persona, avatar, provider,
			parameters, created_at, updated_at
	`
	var paramsRaw []byte
	err = s.DB.QueryRowContext(ctx, query,
		bot.ID, bot.Name, bot.Persona, bot.Avatar,
		bot.Provider, paramsJSON, bot.CreatedAt, bot.UpdatedAt,
	).Scan(
		&bot.ID, &bot.Name, &bot.Persona, &bot.Avatar,
		&bot.Provider, &paramsRaw, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	if len(paramsRaw) > 0 {
		if err := json.Unmarshal(paramsRaw, &bot.Parameters); err != nil {
			return nil, fmt.Errorf("unmarshal parameters: %w", err)
		}
	}

	return bot, nil
}

// GetBot retrieves a bot profile by ID
//
//encore:api public method=GET path=/api/chat/bots/:id
func (s *Service) GetBot(ctx context.Context, id string) (*types.Bot, error) {
	query := `
		SELECT id, name, persona, avatar, provider,
			parameters, created_at, updated_at
		FROM bots WHERE id = $1
	`
	var bot types.Bot
	var paramsRaw []byte
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&bot.ID, &bot.Name, &bot.Persona, &bot.Avatar,
		&bot.Provider, &paramsRaw, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bot not found")
	} else if err != nil {
		return nil, fmt.Errorf("get bot: %w", err)
	}

	if len(paramsRaw) > 0 {
		if err := json.Unmarshal(paramsRaw, &bot.Parameters); err != nil {
			return nil, fmt.Errorf("unmarshal parameters: %w", err)
		}
	}

	return &bot, nil
}

// ListBotsResponse represents the response for listing bots
type ListBotsResponse struct {
	Bots []*types.Bot `json:"bots"`
}

// ListBots retrieves all bot profiles
//
//encore:api public method=GET path=/api/chat/bots
func (s *Service) ListBots(ctx context.Context) (*ListBotsResponse, error) {
	query := `
		SELECT id, name, persona, avatar, provider,
			parameters, created_at, updated_at
		FROM bots
		ORDER BY created_at DESC
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list bots: %w", err)
	}
	defer rows.Close()

	var bots []*types.Bot
	for rows.Next() {
		var bot types.Bot
		var paramsRaw []byte
		err := rows.Scan(
			&bot.ID, &bot.Name, &bot.Persona, &bot.Avatar,
			&bot.Provider, &paramsRaw, &bot.CreatedAt, &bot.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bot: %w", err)
		}

		if len(paramsRaw) > 0 {
			if err := json.Unmarshal(paramsRaw, &bot.Parameters); err != nil {
				return nil, fmt.Errorf("unmarshal parameters: %w", err)
			}
		}

		bots = append(bots, &bot)
	}

	return &ListBotsResponse{Bots: bots}, nil
}

// GetConversation retrieves a conversation by ID
//
//encore:api public method=GET path=/api/chat/conversations/:id
func (s *Service) GetConversation(ctx context.Context, id string) (*types.Conversation, error) {
	query := `
		SELECT id, channel_id, platform, bot_ids,
			created_at, updated_at
		FROM conversations WHERE id = $1
	`
	var conv types.Conversation
	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.ChannelID, &conv.Platform,
		pq.Array(&conv.BotIDs), &conv.CreatedAt, &conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	} else if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	return &conv, nil
}

// ListConversationsResponse represents the response for listing conversations
type ListConversationsResponse struct {
	Conversations []*types.Conversation `json:"conversations"`
}

// ListConversations retrieves all conversations
//
//encore:api public method=GET path=/api/chat/conversations
func (s *Service) ListConversations(ctx context.Context) (*ListConversationsResponse, error) {
	query := `
		SELECT id, channel_id, platform, bot_ids,
			created_at, updated_at
		FROM conversations
		ORDER BY updated_at DESC
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*types.Conversation
	for rows.Next() {
		var conv types.Conversation
		err := rows.Scan(
			&conv.ID, &conv.ChannelID, &conv.Platform,
			pq.Array(&conv.BotIDs), &conv.CreatedAt, &conv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		conversations = append(conversations, &conv)
	}

	return &ListConversationsResponse{Conversations: conversations}, nil
}

// ListMessagesResponse represents the response for listing conversation messages
type ListMessagesResponse struct {
	Messages []*types.Message `json:"messages"`
}

// ListConversationMessages retrieves messages for a conversation
//
//encore:api public method=GET path=/api/chat/conversations/:id/messages
func (s *Service) ListConversationMessages(ctx context.Context, id string) (*ListMessagesResponse, error) {
	// First check if conversation exists
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check conversation existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", id)
	}

	query := `
		SELECT m.id, m.conversation_id, c.channel_id, c.platform,
			m.user_id, m.bot_id, m.content, m.type, m.created_at
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at ASC
	`
	rows, err := s.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("list messages query error: %w", err)
	}
	defer rows.Close()

	var messages []*types.Message
	for rows.Next() {
		var msg types.Message
		var botID sql.NullString // Use sql.NullString to handle NULL bot_id
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.ChannelID,
			&msg.Platform, &msg.UserID, &botID,
			&msg.Content, &msg.Type, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		if botID.Valid {
			msg.BotID = botID.String
		}
		messages = append(messages, &msg)
	}

	return &ListMessagesResponse{Messages: messages}, nil
}

// SendMessage sends a chat message
//
//encore:api public method=POST path=/api/chat/messages
func (s *Service) SendMessage(ctx context.Context, msg *types.Message) (*types.Message, error) {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg_%s", uuid.New().String())
	}
	if msg.UserID == "" {
		msg.UserID = "anonymous"
	}
	if msg.Type == "" {
		msg.Type = "text"
	}
	msg.CreatedAt = time.Now()

	// Start transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get or create conversation
	if msg.ConversationID == "" {
		// Get all available bots
		bots, err := s.ListBots(ctx)
		if err != nil {
			return nil, fmt.Errorf("list bots: %w", err)
		}

		// Extract bot IDs
		botIDs := make([]string, len(bots.Bots))
		for i, bot := range bots.Bots {
			botIDs[i] = bot.ID
		}

		// Set default platform and channel_id if not provided
		channelID := msg.ChannelID
		if channelID == "" {
			channelID = "default"
		}
		platform := msg.Platform
		if platform == "" {
			platform = "local"
		}

		conv := &types.Conversation{
			ID:        fmt.Sprintf("conv_%s", uuid.New().String()),
			ChannelID: channelID,
			Platform:  platform,
			BotIDs:    botIDs,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.CreatedAt,
		}

		query := `
			INSERT INTO conversations (
				id, channel_id, platform, bot_ids,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.ExecContext(ctx, query,
			conv.ID, conv.ChannelID, conv.Platform,
			pq.Array(conv.BotIDs), conv.CreatedAt, conv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("create conversation: %w", err)
		}
		msg.ConversationID = conv.ID
	}

	// Validate conversation exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1)", msg.ConversationID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check conversation existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", msg.ConversationID)
	}

	// Store message
	query := `
		INSERT INTO messages (
			id, conversation_id, user_id, bot_id,
			content, type, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query,
		msg.ID, msg.ConversationID, msg.UserID,
		nil, msg.Content, msg.Type, msg.CreatedAt, // Set bot_id to nil for user messages
	)
	if err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Get the complete message details from database
	query = `
		SELECT m.id, m.conversation_id, c.channel_id, c.platform,
			m.user_id, m.bot_id, m.content, m.type, m.created_at
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE m.id = $1
	`
	var botID sql.NullString
	err = s.DB.QueryRowContext(ctx, query, msg.ID).Scan(
		&msg.ID, &msg.ConversationID, &msg.ChannelID,
		&msg.Platform, &msg.UserID, &botID,
		&msg.Content, &msg.Type, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get message details: %w", err)
	}
	if botID.Valid {
		msg.BotID = botID.String
	}

	// Publish chat event
	event := &types.ChatEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		Type:      "message",
		Platform:  msg.Platform,
		ChannelID: msg.ChannelID,
		Message:   msg,
		Timestamp: msg.CreatedAt,
	}

	_, err = chatpubsub.ChatEvents.Publish(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("publish chat event: %w", err)
	}

	return msg, nil
}

// Initialize subscriptions
var _ = pubsub.NewSubscription(
	chatpubsub.ChatEvents, "process-chat-event",
	pubsub.SubscriptionConfig[*types.ChatEvent]{
		Handler: pubsub.MethodHandler((*Service).processChatEvent),
	},
)

// processChatEvent handles incoming chat events
func (s *Service) processChatEvent(ctx context.Context, event *types.ChatEvent) error {
	// Only process user messages, not bot messages
	if event.Type != "message" || event.Message == nil || event.Message.BotID != "" {
		return nil
	}

	// Get conversation bots
	query := `
		SELECT bot_ids FROM conversations
		WHERE id = $1
	`
	var botIDs []string
	err := s.DB.QueryRowContext(ctx, query, event.Message.ConversationID).Scan(pq.Array(&botIDs))
	if err != nil {
		return fmt.Errorf("get conversation bots: %w", err)
	}

	// Process message for each bot
	for _, botID := range botIDs {
		bot, err := s.GetBot(ctx, botID)
		if err != nil {
			continue
		}

		// Create LLM request
		messages := []llmtypes.Message{
			{
				Role:    "system",
				Content: fmt.Sprintf("You are %s. Respond in character.", bot.Persona),
			},
			{
				Role:    "user",
				Content: event.Message.Content,
			},
		}

		params := llmtypes.Parameters{
			MaxTokens:   bot.Parameters.MaxTokens,
			Temperature: bot.Parameters.Temperature,
		}

		req := &llmtypes.LLMRequestEvent{
			RequestID:      fmt.Sprintf("req_%s", uuid.New().String()),
			BotID:          botID,
			ChannelID:      event.ChannelID,
			ConversationID: event.Message.ConversationID,
			Provider:       bot.Provider,
			Messages:       messages,
			Parameters:     params,
			Timestamp:      time.Now(),
		}

		// Send typing indicator
		typingEvent := &types.ChatEvent{
			EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
			Type:      "typing",
			Platform:  event.Platform,
			ChannelID: event.ChannelID,
			Message: &types.Message{
				BotID: botID,
			},
			Timestamp: time.Now(),
		}
		// Send typing indicator to platform
		if err := s.Broadcast(ctx, &BroadcastRequest{Event: *typingEvent}); err != nil {
			return fmt.Errorf("broadcast typing event: %w", err)
		}

		// Send request to LLM service
		if err := llm.ProcessRequest(ctx, req); err != nil {
			return fmt.Errorf("process llm request: %w", err)
		}
	}

	return nil
}

// Initialize subscription for LLM responses
var _ = pubsub.NewSubscription(
	llmpubsub.GenerationResponses, "handle-llm-response",
	pubsub.SubscriptionConfig[*llmtypes.LLMResponseEvent]{
		Handler: pubsub.MethodHandler((*Service).handleLLMResponse),
	},
)

// handleLLMResponse processes LLM responses and sends them back to the chat
func (s *Service) handleLLMResponse(ctx context.Context, resp *llmtypes.LLMResponseEvent) error {
	if resp.Error != "" {
		return fmt.Errorf("llm error: %s", resp.Error)
	}

	// Get conversation details
	query := `
		SELECT c.id, c.channel_id, c.platform
		FROM conversations c
		JOIN messages m ON m.conversation_id = c.id
		WHERE m.conversation_id = $1
		LIMIT 1
	`
	var conv struct {
		id        string
		channelID string
		platform  string
	}
	err := s.DB.QueryRowContext(ctx, query, resp.ConversationID).Scan(
		&conv.id, &conv.channelID, &conv.platform,
	)
	if err != nil {
		return fmt.Errorf("get conversation: %w", err)
	}

	// Create bot message
	msg := &types.Message{
		ID:             fmt.Sprintf("msg_%s", uuid.New().String()),
		ConversationID: conv.id,
		ChannelID:      conv.channelID,
		Platform:       conv.platform,
		BotID:          resp.BotID,
		Content:        resp.Content,
		Type:           "text",
		CreatedAt:      resp.OccurredAt(),
	}

	// Send message
	_, err = s.SendMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("send bot message: %w", err)
	}

	return nil
}

package local

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/rlog"
	"github.com/google/uuid"

	"encore.app/carriers"
	"encore.app/chat"
	chatpubsub "encore.app/chat/pubsub"
	"encore.app/chat/types"
)

//encore:service
type Service struct{}

// Client represents a connected chat client
type Client struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// Message represents a chat message from the client
type Message struct {
	Content        string `json:"content"`
	ConversationID string `json:"conversation_id"`
}

// Connect establishes a WebSocket connection for the chat
//
//encore:api public method=GET path=/api/local/connect
func (s *Service) Connect(ctx context.Context) (*carriers.WebSocketResponse, error) {
	clientID := fmt.Sprintf("client_%s", uuid.New().String())
	userID := fmt.Sprintf("user_%s", uuid.New().String())

	// Create WebSocket handler
	handler := &wsHandler{
		clientID: clientID,
		userID:   userID,
	}

	return carriers.NewWebSocket(handler), nil
}

// wsHandler implements the WebSocket message handler
type wsHandler struct {
	clientID string
	userID   string
}

// OnMessage handles incoming WebSocket messages
func (h *wsHandler) OnMessage(ctx context.Context, msg []byte) error {
	// Parse message
	var chatMsg Message
	if err := json.Unmarshal(msg, &chatMsg); err != nil {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid message format",
		}
	}

	// Create chat message
	message := &types.Message{
		ID:             fmt.Sprintf("msg_%s", uuid.New().String()),
		ConversationID: chatMsg.ConversationID,
		Platform:       "local",
		ChannelID:      "local",
		UserID:         h.userID,
		Content:        chatMsg.Content,
		Type:           "text",
		CreatedAt:      time.Now(),
	}

	// Send message through chat service
	if _, err := chat.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

// OnClose handles WebSocket connection closure
func (h *wsHandler) OnClose(ctx context.Context) {
	// Cleanup can be added here if needed
}

// ClientID returns the client's unique identifier
func (h *wsHandler) ClientID() string {
	return h.clientID
}

// Initialize subscription for chat events
var _ = pubsub.NewSubscription(
	chatpubsub.ChatEvents, "handle-local-chat",
	pubsub.SubscriptionConfig[*types.ChatEvent]{
		Handler: pubsub.MethodHandler((*Service).handleChatEvent),
	},
)

// handleChatEvent processes chat events and broadcasts them to local clients
func (s *Service) handleChatEvent(ctx context.Context, event *types.ChatEvent) error {
	// Only process local platform events
	if event.Platform != "local" {
		return nil
	}

	// Log event details for debugging
	rlog.Info("handling chat event",
		"event_type", event.Type,
		"channel_id", event.ChannelID,
		"message_id", event.Message.ID,
		"conversation_id", event.Message.ConversationID,
	)

	// Broadcast message to all connected clients
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := carriers.Broadcast(msg); err != nil {
		return fmt.Errorf("broadcast message: %w", err)
	}

	return nil
}

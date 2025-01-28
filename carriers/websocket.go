package carriers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"encore.dev/rlog"
	"github.com/gorilla/websocket"
)

// WebSocketHandler defines the interface for handling WebSocket events
type WebSocketHandler interface {
	OnMessage(ctx context.Context, msg []byte) error
	OnClose(ctx context.Context)
	ClientID() string
}

// WebSocketResponse represents the response for a WebSocket connection
type WebSocketResponse struct {
	ClientID string `json:"client_id"`
}

// wsClient represents a connected WebSocket client
type wsClient struct {
	id      string
	handler WebSocketHandler
	send    chan []byte
}

var (
	clients   = make(map[string]*wsClient)
	clientsMu sync.RWMutex
)

// NewWebSocket creates a new WebSocket connection and registers the client
func NewWebSocket(handler WebSocketHandler) *WebSocketResponse {
	clientID := handler.ClientID()

	// Register client
	clientsMu.Lock()
	clients[clientID] = &wsClient{
		id:      clientID,
		handler: handler,
		send:    make(chan []byte, 256),
	}
	clientsMu.Unlock()

	return &WebSocketResponse{
		ClientID: clientID,
	}
}

// Broadcast sends a message to all connected clients
func Broadcast(msg []byte) error {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	var lastErr error
	for _, client := range clients {
		select {
		case client.send <- msg:
			// Message sent successfully
		default:
			err := fmt.Errorf("failed to send message to client %s: buffer full", client.id)
			rlog.Error("failed to send message to client", "client_id", client.id, "error", err)
			lastErr = err
		}
	}

	return lastErr
}

// BroadcastEvent sends a message to all connected clients for a specific channel
func BroadcastEvent(msg []byte, channelID string) error {
	// For now, just use Broadcast since we're using a single channel
	// In the future, we can add channel-specific routing
	return Broadcast(msg)
}

// HandleWebSocket upgrades an HTTP connection to WebSocket and manages the connection
//
//encore:api public raw method=GET path=/api/ws/:client_id
func HandleWebSocket(w http.ResponseWriter, req *http.Request) {
	// Extract client ID from URL path
	clientID := req.URL.Path[len("/api/ws/"):]
	if clientID == "" {
		http.Error(w, "missing client_id", http.StatusBadRequest)
		return
	}

	// Validate client ID
	clientsMu.RLock()
	client, exists := clients[clientID]
	clientsMu.RUnlock()
	if !exists {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	// Upgrade connection to WebSocket
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		rlog.Error("failed to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer c.Close()

	// Start client goroutines
	done := make(chan struct{})
	go func() {
		defer close(done)
		readMessages(req.Context(), c, client.handler)
	}()
	writeMessages(c, client.send, done)

	// Cleanup client on disconnect
	clientsMu.Lock()
	delete(clients, clientID)
	close(client.send)
	clientsMu.Unlock()
}

// readMessages reads messages from the WebSocket connection
func readMessages(ctx context.Context, c *websocket.Conn, handler WebSocketHandler) {
	defer func() {
		c.Close()
		handler.OnClose(ctx)
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				rlog.Error("websocket error", "error", err)
			}
			break
		}

		if err := handler.OnMessage(ctx, message); err != nil {
			rlog.Error("handle message error", "error", err)
		}
	}
}

// writeMessages writes messages to the WebSocket connection
func writeMessages(c *websocket.Conn, send <-chan []byte, done <-chan struct{}) {
	defer c.Close()

	for {
		select {
		case message, ok := <-send:
			if !ok {
				c.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
				rlog.Error("write message error", "error", err)
				return
			}
		case <-done:
			return
		}
	}
}

// upgrader configures WebSocket connection parameters
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

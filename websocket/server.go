package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"encore.dev/rlog"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// BroadcastParams represents the parameters for broadcasting a status update
type BroadcastParams struct {
	Status  string  `json:"status"`
	ID      string  `json:"id"`
	Message *string `json:"message,omitempty"`
}

// BroadcastResponse represents the response from a broadcast operation
type BroadcastResponse struct {
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
}

// Websocket handles WebSocket connections
//
//encore:api public raw path=/ws
func (s *Service) Websocket(w http.ResponseWriter, req *http.Request) {
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		rlog.Error("could not upgrade websocket", "err", err)
		return
	}

	// Register client
	s.mu.Lock()
	s.clients[c] = true
	s.mu.Unlock()

	// Ensure we unregister client on disconnect
	defer func() {
		s.mu.Lock()
		delete(s.clients, c)
		s.mu.Unlock()
		c.Close()
	}()

	// Handle incoming messages
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				rlog.Error("websocket error", "err", err)
			}
			break
		}
		rlog.Info("received message", "msg", string(message))

		// Broadcast message to all clients
		s.mu.RLock()
		for client := range s.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				rlog.Error("error broadcasting message", "err", err)
			}
		}
		s.mu.RUnlock()
	}
}

// Broadcast sends a status update to all connected clients
//
//encore:api public method=POST path=/broadcast
func (s *Service) Broadcast(ctx context.Context, params *BroadcastParams) (*BroadcastResponse, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	msg := struct {
		Type    string `json:"type"`
		Content struct {
			ID        string  `json:"id"`
			Status    string  `json:"status"`
			Timestamp string  `json:"timestamp"`
			Message   *string `json:"message,omitempty"`
		} `json:"content"`
	}{
		Type: "status_update",
		Content: struct {
			ID        string  `json:"id"`
			Status    string  `json:"status"`
			Timestamp string  `json:"timestamp"`
			Message   *string `json:"message,omitempty"`
		}{
			ID:        params.ID,
			Status:    params.Status,
			Timestamp: now,
			Message:   params.Message,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("error marshaling message: %v", err)
	}

	// Broadcast to all clients
	s.mu.RLock()
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			rlog.Error("error broadcasting message", "err", err)
		}
	}
	s.mu.RUnlock()

	return &BroadcastResponse{
		Success:   true,
		Timestamp: now,
	}, nil
}

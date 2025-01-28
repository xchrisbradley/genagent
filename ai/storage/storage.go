package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/xchrisbradley/genagent/plugins/ai/llm"
)

// MessageStore handles persistence of chat messages
type MessageStore struct {
	filePath string
}

// storedMessage adds serialization tags to the Message type
type storedMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMessageStore creates a new message store
func NewMessageStore(storageDir string) (*MessageStore, error) {
	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, err
	}

	return &MessageStore{
		filePath: filepath.Join(storageDir, "chat_history.json"),
	}, nil
}

// LoadMessages loads messages from storage
func (s *MessageStore) LoadMessages() ([]llm.Message, error) {
	// If file doesn't exist, return empty slice
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []llm.Message{}, nil
	}

	// Read file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	// Parse messages
	var storedMessages []storedMessage
	if err := json.Unmarshal(data, &storedMessages); err != nil {
		return nil, err
	}

	// Convert to llm.Message format
	messages := make([]llm.Message, len(storedMessages))
	for i, msg := range storedMessages {
		messages[i] = llm.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
	}

	return messages, nil
}

// SaveMessages saves messages to storage
func (s *MessageStore) SaveMessages(messages []llm.Message) error {
	// Convert to stored format
	storedMessages := make([]storedMessage, len(messages))
	for i, msg := range messages {
		storedMessages[i] = storedMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
	}

	// Convert messages to JSON
	data, err := json.MarshalIndent(storedMessages, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(s.filePath, data, 0644)
}

// ClearHistory deletes the chat history file
func (s *MessageStore) ClearHistory() error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(s.filePath)
}

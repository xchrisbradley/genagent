package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

type MessageStore struct {
	dir   string
	mutex sync.RWMutex
}

func NewMessageStore(dir string) (*MessageStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	return &MessageStore{dir: dir}, nil
}

func (s *MessageStore) SaveMessages(messages []Message) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	filePath := filepath.Join(s.dir, "messages.json")
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write messages file: %v", err)
	}

	return nil
}

func (s *MessageStore) LoadMessages() ([]Message, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	filePath := filepath.Join(s.dir, "messages.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, fmt.Errorf("failed to read messages file: %v", err)
	}

	var messages []Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %v", err)
	}

	return messages, nil
}

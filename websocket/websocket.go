package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

//encore:service
type Service struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func initService() (*Service, error) {
	return &Service{
		clients: make(map[*websocket.Conn]bool),
	}, nil
}

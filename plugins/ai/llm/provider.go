package llm

import (
	"context"
	"time"
)

type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

type Provider interface {
	Name() string
	Process(ctx context.Context, messages []Message) (string, error)
}

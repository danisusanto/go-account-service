package event

import (
	"context"
	"time"
)

type Payload struct {
	EventID   int       `json:"event_id"`
	EventName string    `json:"event_name"`
	CreatedAt time.Time `json:"created_at"`
	Data      any       `json:"payload"`
}

type Event interface {
	Publish(ctx context.Context, payload *Payload) error
	// Consume() error
}

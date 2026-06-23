package event

import (
	"fmt"

	"github.com/dani-susanto/go-account-service/internal/domain/event"
)

type Handler struct {
	handlers EventHandlers
}

type EventHandlers map[string]func(event.Payload) error

func NewEventHandler() *Handler {
	return &Handler{
		handlers: make(EventHandlers),
	}
}

func (d *Handler) Exec(eventName string, payload event.Payload) error {
	handler, ok := d.handlers[eventName]
	if !ok {
		return fmt.Errorf("no handler registered for event: %s", eventName)
	}

	if err := handler(payload); err != nil {
		return err
	}

	return nil
}

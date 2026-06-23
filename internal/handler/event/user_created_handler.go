package event

import (
	"context"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
)

type userCreatedHandler struct {
	handler     *Handler
	userCreated usecase.UserCreatedUsecase
	json        json.JSON
	log         log.Log
}

func NewUserCreatedHandler(
	handler *Handler,
	userCreated usecase.UserCreatedUsecase,
	json json.JSON,
	log log.Log,
) *userCreatedHandler {
	return &userCreatedHandler{
		handler:     handler,
		userCreated: userCreated,
		json:        json,
		log:         log,
	}
}

func (h *userCreatedHandler) Register(event string) *userCreatedHandler {
	h.handler.handlers[event] = h.UserCreatedHandler
	return h
}

func (h *userCreatedHandler) Transform(payload event.Payload) *event.UserCreated {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserCreated
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userCreatedHandler) UserCreatedHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userCreated.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserCreatedEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

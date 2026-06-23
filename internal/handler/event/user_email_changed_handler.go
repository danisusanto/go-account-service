package event

import (
	"context"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
)

type userEmailChangedHandler struct {
	handler          *Handler
	userEmailChanged usecase.UserEmailChangedUsecase
	json             json.JSON
	log              log.Log
}

func NewUserEmailChangedHandler(
	handler *Handler,
	userEmailChanged usecase.UserEmailChangedUsecase,
	json json.JSON,
	log log.Log,
) *userEmailChangedHandler {
	return &userEmailChangedHandler{
		handler:          handler,
		userEmailChanged: userEmailChanged,
		json:             json,
		log:              log,
	}
}

func (h *userEmailChangedHandler) Register(event string) *userEmailChangedHandler {
	h.handler.handlers[event] = h.UserEmailChangedHandler
	return h
}

func (h *userEmailChangedHandler) Transform(payload event.Payload) *event.UserEmailChanged {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserEmailChanged
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userEmailChangedHandler) UserEmailChangedHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userEmailChanged.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserEmailChangedEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

package event

import (
	"context"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
)

type userUpdatedHandler struct {
	handler     *Handler
	userUpdated usecase.UserUpdatedUsecase
	json        json.JSON
	log         log.Log
}

func NewUserUpdatedHandler(handler *Handler, userUpdated usecase.UserUpdatedUsecase, json json.JSON, log log.Log) *userUpdatedHandler {
	return &userUpdatedHandler{
		handler:     handler,
		userUpdated: userUpdated,
		json:        json,
		log:         log,
	}
}

func (h *userUpdatedHandler) Register(event string) *userUpdatedHandler {
	h.handler.handlers[event] = h.UserUpdatedHandler
	return h
}

func (h *userUpdatedHandler) Transform(payload event.Payload) *event.UserUpdated {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserUpdated
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userUpdatedHandler) UserUpdatedHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userUpdated.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserUpdatedEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

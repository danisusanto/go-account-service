package event

import (
	"context"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
)

type userVerificationCreatedHandler struct {
	handler                 *Handler
	userVerificationCreated usecase.UserVerificationCreatedUsecase
	json                    json.JSON
	log                     log.Log
}

func NewUserVerificationCreatedHandler(
	handler *Handler,
	userVerificationCreated usecase.UserVerificationCreatedUsecase,
	json json.JSON,
	log log.Log,
) *userVerificationCreatedHandler {
	return &userVerificationCreatedHandler{
		handler:                 handler,
		userVerificationCreated: userVerificationCreated,
		json:                    json,
		log:                     log,
	}
}

func (h *userVerificationCreatedHandler) Register(event string) *userVerificationCreatedHandler {
	h.handler.handlers[event] = h.UserVerificationCreatedHandler
	return h
}

func (h *userVerificationCreatedHandler) Transform(payload event.Payload) *event.UserVerificationCreated {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserVerificationCreated
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userVerificationCreatedHandler) UserVerificationCreatedHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userVerificationCreated.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserVerificationCreatedEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

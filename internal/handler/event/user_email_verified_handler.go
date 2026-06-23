package event

import (
	"context"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
)

type userEmailVerifiedHandler struct {
	handler           *Handler
	userEmailVerified usecase.UserEmailVerifiedUsecase
	json              json.JSON
	log               log.Log
}

func NewUserEmailVerifiedHandler(
	handler *Handler,
	userEmailVerified usecase.UserEmailVerifiedUsecase,
	json json.JSON,
	log log.Log,
) *userEmailVerifiedHandler {
	return &userEmailVerifiedHandler{
		handler:           handler,
		userEmailVerified: userEmailVerified,
		json:              json,
		log:               log,
	}
}

func (h *userEmailVerifiedHandler) Register(event string) *userEmailVerifiedHandler {
	h.handler.handlers[event] = h.UserEmailVerifiedHandler
	return h
}

func (h *userEmailVerifiedHandler) Transform(payload event.Payload) *event.UserEmailVerified {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserEmailVerified
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userEmailVerifiedHandler) UserEmailVerifiedHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userEmailVerified.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserEmailVerifiedEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

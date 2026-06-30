package event

import (
	"context"

	"github.com/danisusanto/go-common/log"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-common/json"
)

type userResetPasswordHandler struct {
	handler           *Handler
	userResetPassword usecase.UserResetPasswordUsecase
	json              json.JSON
	log               log.Log
}

func NewUserResetPasswordHandler(
	handler *Handler,
	userResetPassword usecase.UserResetPasswordUsecase,
	json json.JSON,
	log log.Log,
) *userResetPasswordHandler {
	return &userResetPasswordHandler{
		handler:           handler,
		userResetPassword: userResetPassword,
		json:              json,
		log:               log,
	}
}

func (h *userResetPasswordHandler) Register(event string) *userResetPasswordHandler {
	h.handler.handlers[event] = h.UserResetPasswordHandler
	return h
}

func (h *userResetPasswordHandler) Transform(payload event.Payload) *event.UserResetPassword {
	byteData, err := h.json.Encode(payload.Data)
	if err != nil {
		h.log.Error(err.Error())
	}

	var data event.UserResetPassword
	h.json.Decode([]byte(byteData), &data)
	return &data
}

func (h *userResetPasswordHandler) UserResetPasswordHandler(payload event.Payload) error {
	ctx := context.Background()
	if err := h.userResetPassword.Exec(ctx, h.Transform(payload)); err != nil {
		h.log.Error(
			event.UserResetPasswordEvent,
			"error", err,
			"payload", payload,
		)
		return err
	}
	return nil
}

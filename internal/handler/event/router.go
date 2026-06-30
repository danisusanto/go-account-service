package event

import (
	domainevent "github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/provider"
	"github.com/danisusanto/go-common/json"
	"github.com/danisusanto/go-common/log"
)

func (h *Handler) SetupEventRoutes(
	usecase *provider.Usecases,
	json json.JSON,
	log log.Log,
) {
	// NOTE : If using RabbitMQ, don't forget to register binding routes first in the rabbitmq provider

	NewUserCreatedHandler(h, usecase.UserCreated, json, log).Register(domainevent.UserCreatedEvent)
	NewUserUpdatedHandler(h, usecase.UserUpdated, json, log).Register(domainevent.UserUpdatedEvent)
	NewUserEmailVerifiedHandler(h, usecase.UserEmailVerified, json, log).Register(domainevent.UserEmailVerifiedEvent)
	NewUserResetPasswordHandler(h, usecase.UserResetPassword, json, log).Register(domainevent.UserResetPasswordEvent)
	NewUserVerificationCreatedHandler(h, usecase.UserVerificationCreated, json, log).Register(domainevent.UserVerificationCreatedEvent)
	NewUserEmailChangedHandler(h, usecase.UserEmailChanged, json, log).Register(domainevent.UserEmailChangedEvent)
}

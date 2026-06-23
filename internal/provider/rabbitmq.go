package provider

import (
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func RabbitMQRoutes() []rabbitmq.Routes {
	return []rabbitmq.Routes{
		{Queue: "email_queue", Route: event.UserCreatedEvent},
		{Queue: "email_queue", Route: event.UserVerificationCreatedEvent},
		{Queue: "email_queue", Route: event.UserEmailVerifiedEvent},
		{Queue: "email_queue", Route: event.UserResetPasswordEvent},
		{Queue: "email_queue", Route: event.UserEmailChangedEvent},
		{Queue: "update_queue", Route: event.UserUpdatedEvent},
	}
}

func SetupRabbitMQ(env env.Env) (*amqp.Channel, error) {
	channel, err := rabbitmq.New(
		env.Get("RABBITMQ_USERNAME"),
		env.Get("RABBITMQ_PASSWORD"),
		env.Get("RABBITMQ_HOST"),
		env.Get("RABBITMQ_PORT"),
	)
	if err != nil {
		return nil, err
	}

	exchangeName := env.Get("RABBITMQ_EXCHANGE")
	err = rabbitmq.Exchange(
		channel,
		exchangeName,
		env.Get("RABBITMQ_EXCHANGE_TYPE"),
	)
	if err != nil {
		return nil, err
	}

	err = rabbitmq.Queue(
		channel,
		exchangeName,
		RabbitMQRoutes(),
	)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

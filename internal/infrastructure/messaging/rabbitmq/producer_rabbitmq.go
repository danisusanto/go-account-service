package rabbitmq

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-common/json"
	"github.com/dani-susanto/go-common/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type producer struct {
	channel  *amqp.Channel
	exchange string
	json     json.JSON
	log      log.Log
}

func NewProducerRabbitMQ(
	channel *amqp.Channel,
	exchange string,
	json json.JSON,
	log log.Log,
) event.Event {
	return &producer{
		channel:  channel,
		exchange: exchange,
		json:     json,
		log:      log,
	}
}

func (r *producer) Publish(ctx context.Context, payload *event.Payload) error {
	data, err := r.json.Encode(payload)
	if err != nil {
		return err
	}

	r.log.Info(
		"publishing event",
		"event_name", payload.EventName,
		"payload", data,
	)

	err = r.channel.PublishWithContext(ctx,
		r.exchange,        // exchange
		payload.EventName, // routing key (event)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

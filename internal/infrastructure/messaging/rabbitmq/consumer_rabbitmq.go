package rabbitmq

import (
	domainevent "github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/handler/event"
	"github.com/danisusanto/go-account-service/internal/provider"
	"github.com/danisusanto/go-common/json"
	"github.com/danisusanto/go-common/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type consumer struct {
	channel *amqp.Channel
	handler *event.Handler
	log     log.Log
	json    json.JSON
}

func NewConsumerRabbitMQ(
	channel *amqp.Channel,
	handler *event.Handler,
	log log.Log,
	json json.JSON,
) *consumer {
	return &consumer{
		channel: channel,
		handler: handler,
		log:     log,
		json:    json,
	}
}

func (r *consumer) Consume() error {
	routes := provider.RabbitMQRoutes()

	for _, v := range routes {
		msgs, err := r.channel.Consume(
			v.Queue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		go func(queue string, msgs <-chan amqp.Delivery) {
			for msg := range msgs {
				var payload domainevent.Payload
				_ = r.json.Decode(msg.Body, &payload)
				if err := r.handler.Exec(msg.RoutingKey, payload); err != nil {
					r.log.Error(
						"consuming event message failed",
						"error", err,
						"event_name", msg.RoutingKey,
						"payload", payload,
					)
				}
			}
		}(v.Queue, msgs)
	}

	select {}
}

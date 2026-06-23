package bootstrap

import (
	"github.com/dani-susanto/go-account-service/internal/handler/event"
	"github.com/dani-susanto/go-account-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/dani-susanto/go-account-service/internal/provider"
	"github.com/dani-susanto/go-common/hash"
	"github.com/dani-susanto/go-common/json"
	"github.com/dani-susanto/go-common/log"
	"github.com/dani-susanto/go-common/postgres"
	"github.com/dani-susanto/go-common/redis"
	amqp "github.com/rabbitmq/amqp091-go"
)

type eventServer struct {
	channel *amqp.Channel
	usecase *provider.Usecases
	log     log.Log
	json    json.JSON
}

func NewEventServer() (*eventServer, error) {
	// Setup Environment
	env, err := provider.SetupEnv()
	if err != nil {
		return nil, err
	}

	// Setup Log
	log := log.New()

	// Setup JSON
	json := json.New()

	// Setup Messaging Infrastructure (RabbitMQ)
	messaging, err := provider.SetupRabbitMQ(env)
	if err != nil {
		return nil, err
	}

	// Setup Database Infrastructure (Postgres)
	db, err := postgres.New(
		env.Get("DB_USERNAME"),
		env.Get("DB_PASSWORD"),
		env.Get("DB_HOST"),
		env.Get("DB_PORT"),
		env.Get("DB_DATABASE"),
	)
	if err != nil {
		return nil, err
	}

	// Setup Cache Infrastructure (Redis)
	cache, err := redis.New(
		env.Get("REDIS_USERNAME"),
		env.Get("REDIS_PASSWORD"),
		env.Get("REDIS_HOST"),
		env.Get("REDIS_PORT"),
		env.Get("REDIS_DATABASE"),
	)
	if err != nil {
		return nil, err
	}

	// Setup Domain Repository Implementation
	repository := provider.SetupRepository(db, cache, log)

	// Setup Domain Event Implementation (RabbitMQ)
	event := rabbitmq.NewProducerRabbitMQ(messaging, env.Get("RABBITMQ_EXCHANGE"), json, log)

	// Setup Email Provider
	email, err := provider.SetupSMTP(env)
	if err != nil {
		return nil, err
	}

	// Setup Application Usecase
	usecase := provider.SetupEventUsecase(
		repository,
		event,
		email,
		env,
		hash.New(),
		log,
	)

	return &eventServer{
		channel: messaging,
		usecase: usecase,
		log:     log,
		json:    json,
	}, nil
}

func (server *eventServer) Run() error {
	handler := event.NewEventHandler()
	handler.SetupEventRoutes(server.usecase, server.json, server.log)

	server.log.Info("starting account service event")
	server.log.Info("ready to handle connections")
	consumer := rabbitmq.NewConsumerRabbitMQ(server.channel, handler, server.log, server.json)
	if err := consumer.Consume(); err != nil {
		server.log.Error("server failed: %s", err)
		return err
	}
	return nil
}

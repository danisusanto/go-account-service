package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dani-susanto/go-account-service/internal/handler/rest"
	"github.com/dani-susanto/go-account-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/dani-susanto/go-account-service/internal/provider"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/hash"
	"github.com/dani-susanto/go-common/json"
	"github.com/dani-susanto/go-common/jwt"
	"github.com/dani-susanto/go-common/log"
	"github.com/dani-susanto/go-common/postgres"
	"github.com/dani-susanto/go-common/redis"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/dani-susanto/go-common/validator"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type restServer struct {
	usecase        *provider.Usecases
	cache          *goredis.Client
	env            env.Env
	accessTokenJWT jwt.JWT
	log            log.Log
	json           json.JSON
}

func NewRestServer() (*restServer, error) {
	// Setup Environment
	env, err := provider.SetupEnv()
	if err != nil {
		return nil, fmt.Errorf("ENV %s", err)
	}

	// Setup Telemetry
	ctx := context.Background()
	tracer, err := provider.SetupTelemetry(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("Telemetry %s", err)
	}

	// Setup Log
	log := log.New()

	// Setup JSON
	json := json.New()

	// Setup Messaging Infrastructure (RabbitMQ)
	messaging, err := provider.SetupRabbitMQ(env)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ %s", err)
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
		return nil, fmt.Errorf("Postgres %s", err)
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
		return nil, fmt.Errorf("Redis %s", err)
	}

	// Setup Domain Event Implementation (RabbitMQ)
	event := rabbitmq.NewProducerRabbitMQ(messaging, env.Get("RABBITMQ_EXCHANGE"), json, log)

	// Setup Domain Repository Implementation
	repository := provider.SetupRepository(db, cache, log)

	// Setup JWT
	accessTokenJWT, err := provider.SetupAccessTokenJWT(env)
	if err != nil {
		return nil, err
	}
	refreshTokenJWT, err := provider.SetupRefreshTokenJWT(env)
	if err != nil {
		return nil, err
	}

	// Setup Application Usecase
	usecase := provider.SetupUsecase(
		repository,
		event,
		tracer,
		accessTokenJWT,
		refreshTokenJWT,
		env,
		hash.New(),
		log,
	)

	return &restServer{
		usecase:        usecase,
		cache:          cache,
		env:            env,
		accessTokenJWT: accessTokenJWT,
		log:            log,
		json:           json,
	}, nil
}

func (s *restServer) Run() error {
	responder := commonrest.NewResponder(s.json)

	globalRateLimiterAttempt, err := strconv.Atoi(s.env.Get("APP_RATE_LIMITER_ATTEMPT"))
	if err != nil {
		s.log.Error("failed to parse global rate limiter attempt", "error", err.Error())
	}

	router := commonrest.New(
		http.NewServeMux(),
		s.cache,
		globalRateLimiterAttempt,
		s.env.Get("APP_RATE_LIMITER_DURATION"),
		s.accessTokenJWT,
		responder,
		s.log,
	)

	rest.SetupHTTPRoutes(
		"/api/v1",
		s.env,
		s.usecase,
		router,
		validator.New(),
		s.json,
		responder,
		s.log,
	)

	port := ":8001"
	server := http.Server{
		Addr: port,
		Handler: otelhttp.NewHandler(
			router.Log(
				router.GlobalRateLimiter(
					router.GetMux(),
				),
			), "rest-handler",
		),
	}

	s.log.Info("starting account service rest")
	s.log.Info("listening", "port", port)
	s.log.Info("ready to handle connections")

	if err := server.ListenAndServe(); err != nil {
		s.log.Error("server failed", "error", err)
		return err
	}
	return nil
}

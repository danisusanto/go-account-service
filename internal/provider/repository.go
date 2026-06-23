package provider

import (
	"database/sql"

	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-account-service/internal/infrastructure/cache/redis"
	"github.com/dani-susanto/go-account-service/internal/infrastructure/database/postgres"
	"github.com/dani-susanto/go-common/log"
	goredis "github.com/redis/go-redis/v9"
)

type repositories struct {
	userRepository             repository.UserRepository
	userVerificationRepository repository.UserVerificationRepository
	userSessionRepository      repository.UserSessionRepository
}

func SetupRepository(
	db *sql.DB,
	cache *goredis.Client,
	log log.Log,
) *repositories {
	userRepositoryPostgres := postgres.NewUserRepositoryPostgres(db, log)
	userVerificationRepositoryPostgres := postgres.NewUserVerificationRepositoryPostgres(db, log)
	userSessionRepository := redis.NewUserSessionRepository(cache, log)

	return &repositories{
		userRepository:             userRepositoryPostgres,
		userVerificationRepository: userVerificationRepositoryPostgres,
		userSessionRepository:      userSessionRepository,
	}
}

package redis

import (
	"context"
	"time"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/log"
	"github.com/redis/go-redis/v9"
)

type userSessionRepository struct {
	redis *redis.Client
	log   log.Log
}

func NewUserSessionRepository(
	redis *redis.Client,
	log log.Log,
) repository.UserSessionRepository {
	return &userSessionRepository{
		redis: redis,
		log:   log,
	}
}

func (r *userSessionRepository) Create(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error) {
	key := entity.NewSessionKey(session.SessionType, session.SessionID)

	err := r.redis.HSet(ctx, key, session).Err()
	if err != nil {
		return nil, err
	}

	if session.ExpiredAt != nil {
		ttl := time.Until(*session.ExpiredAt)
		if ttl > 0 {
			err = r.redis.Expire(ctx, key, ttl).Err()
			if err != nil {
				return nil, err
			}
		}
	}

	return session, nil
}

func (r *userSessionRepository) CreateBulk(ctx context.Context, sessions []*entity.UserSession) ([]*entity.UserSession, error) {
	for _, v := range sessions {
		r.Create(ctx, v)
	}
	return sessions, nil
}

func (r *userSessionRepository) FindByKey(ctx context.Context, key string) (*entity.UserSession, error) {

	cmd := r.redis.HGetAll(ctx, key)

	result, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, repository.ErrNotFound
	}

	session := new(entity.UserSession)

	err = cmd.Scan(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *userSessionRepository) DeleteByKey(ctx context.Context, key string) error {
	return r.redis.Del(ctx, key).Err()
}

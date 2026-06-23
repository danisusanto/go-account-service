package repository

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
)

type UserSessionRepository interface {
	Create(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error)
	CreateBulk(ctx context.Context, sessions []*entity.UserSession) ([]*entity.UserSession, error)
	FindByKey(ctx context.Context, key string) (*entity.UserSession, error)
	DeleteByKey(ctx context.Context, key string) error
}

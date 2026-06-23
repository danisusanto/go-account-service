package repository

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)

	Get(ctx context.Context, id int) (*entity.User, error)
	List(ctx context.Context) ([]entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	Update(ctx context.Context, id int, user *entity.User) (*entity.User, error)
	MarkEmailAsVerified(ctx context.Context, user *entity.User) error
	SetPassword(ctx context.Context, id int, password string) error
	SetEmail(ctx context.Context, id int, email string) error

	Delete(ctx context.Context, id int) error
}

package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type GetUserUsecase interface {
	Exec(ctx context.Context, userID int) (*GetUserUsecaseOutput, error)
}

type getUserUsecase struct {
	userRepository repository.UserRepository
	log            log.Log
}

type GetUserUsecaseOutput struct {
	ID          int         `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Email       string      `json:"email,omitempty"`
	CountryCode string      `json:"country_code,omitempty"`
	PhoneNumber string      `json:"phone_number,omitempty"`
	GoogleID    string      `json:"google_id,omitempty"`
	IsActive    bool        `json:"is_active,omitempty"`
	Role        entity.Role `json:"role,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
}

func NewGetUserUsecase(
	userRepository repository.UserRepository,
	log log.Log,
) GetUserUsecase {
	return &getUserUsecase{
		userRepository: userRepository,
		log:            log,
	}
}

func (u *getUserUsecase) Exec(ctx context.Context, userID int) (*GetUserUsecaseOutput, error) {
	user, err := u.userRepository.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	return &GetUserUsecaseOutput{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		CountryCode: user.CountryCode,
		PhoneNumber: user.PhoneNumber,
		GoogleID:    user.GoogleID,
		IsActive:    user.IsUserActive(),
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

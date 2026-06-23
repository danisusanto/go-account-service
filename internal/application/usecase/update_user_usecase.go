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

type UpdateUserUsecase interface {
	Exec(ctx context.Context, i *UpdateUserUsecaseInput) (*UpdateUserUsecaseOutput, error)
}

type updateUserUsecase struct {
	userRepository repository.UserRepository
	log            log.Log
}

type UpdateUserUsecaseInput struct {
	UserID      int
	Name        string
	CountryCode string
	PhoneNumber string
	IsActive    *bool
	Role        entity.Role
	UpdatedBy   int
}

type UpdateUserUsecaseOutput struct {
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

func NewUpdateUserUsecase(
	userRepository repository.UserRepository,
	log log.Log,
) UpdateUserUsecase {
	return &updateUserUsecase{
		userRepository: userRepository,
		log:            log,
	}
}

func (u *updateUserUsecase) Exec(ctx context.Context, i *UpdateUserUsecaseInput) (*UpdateUserUsecaseOutput, error) {
	_, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	user, err := u.userRepository.Update(
		ctx,
		i.UserID,
		&entity.User{
			Name:        i.Name,
			CountryCode: i.CountryCode,
			PhoneNumber: i.PhoneNumber,
			IsActive:    i.IsActive,
			Role:        i.Role,
			UpdatedBy:   i.UpdatedBy,
		},
	)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	return &UpdateUserUsecaseOutput{
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

package usecase

import (
	"context"
	"errors"

	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/hash"
	"github.com/danisusanto/go-common/log"
)

type ChangePasswordUsecase interface {
	Exec(ctx context.Context, i *ChangePasswordUsecaseInput) error
}

type changePasswordUsecase struct {
	userRepository repository.UserRepository
	hash           hash.Hash
	log            log.Log
}

type ChangePasswordUsecaseInput struct {
	UserID          int
	CurrentPassword string
	NewPassword     string
}

func NewChangePasswordUsecase(
	userRepository repository.UserRepository,
	hash hash.Hash,
	log log.Log,
) ChangePasswordUsecase {
	return &changePasswordUsecase{
		userRepository: userRepository,
		hash:           hash,
		log:            log,
	}
}

func (u *changePasswordUsecase) Exec(ctx context.Context, i *ChangePasswordUsecaseInput) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if !u.hash.Compare(i.CurrentPassword, user.Password) {
		return exception.Throw(exception.BAD_REQUEST, "the current password is incorrect")
	}

	password, err := u.hash.Hash(i.NewPassword)
	if err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.userRepository.SetPassword(ctx, user.ID, password); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	return nil

}

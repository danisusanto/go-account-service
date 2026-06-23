package usecase

import (
	"context"
	"errors"

	"github.com/dani-susanto/go-common/hash"

	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-account-service/internal/domain/notification/email"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type UserEmailChangedUsecase interface {
	Exec(ctx context.Context, i *event.UserEmailChanged) error
}

type userEmailChangedUsecase struct {
	userRepository repository.UserRepository
	email          email.Email
	env            env.Env
	hash           hash.Hash
	log            log.Log
}

func NewUserEmailChangedUsecase(
	userRepository repository.UserRepository,
	email email.Email,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) UserEmailChangedUsecase {
	return &userEmailChangedUsecase{
		userRepository: userRepository,
		email:          email,
		env:            env,
		hash:           hash,
		log:            log,
	}
}

func (u *userEmailChangedUsecase) Exec(ctx context.Context, i *event.UserEmailChanged) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	u.log.Info("sending change email success", "email", user.Email)
	if err := u.email.Send(
		ctx,
		email.ChangeEmailSuccessSubject,
		user.Email,
		&email.Payload{
			TemplateName: email.ChangeEmailSuccessTemplate,
			Data: &email.ChangeEmailSuccess{
				Name:     user.Name,
				LoginURL: u.env.Get("APP_URL"),
			},
		},
	); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send change email success")
	}
	u.log.Info("change email success sent", "email", user.Email)
	return nil
}

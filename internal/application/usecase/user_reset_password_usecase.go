package usecase

import (
	"context"
	"errors"

	"github.com/danisusanto/go-common/hash"

	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type UserResetPasswordUsecase interface {
	Exec(ctx context.Context, i *event.UserResetPassword) error
}

type userResetPasswordUsecase struct {
	userRepository repository.UserRepository
	email          email.Email
	env            env.Env
	hash           hash.Hash
	log            log.Log
}

func NewUserResetPasswordUsecase(
	userRepository repository.UserRepository,
	email email.Email,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) UserResetPasswordUsecase {
	return &userResetPasswordUsecase{
		userRepository: userRepository,
		email:          email,
		env:            env,
		hash:           hash,
		log:            log,
	}
}

func (u *userResetPasswordUsecase) Exec(ctx context.Context, i *event.UserResetPassword) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	u.log.Info("sending reset password success email", "email", user.Email)
	if err := u.email.Send(
		ctx,
		email.ResetPasswordSuccessSubject,
		user.Email,
		&email.Payload{
			TemplateName: email.ResetPasswordSuccessTemplate,
			Data: &email.ResetPasswordSuccess{
				Name:     user.Name,
				LoginURL: u.env.Get("APP_URL"),
			},
		},
	); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send reset password success email")
	}
	u.log.Info("reset password success email sent", "email", user.Email)
	return nil
}

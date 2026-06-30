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

type UserEmailVerifiedUsecase interface {
	Exec(ctx context.Context, i *event.UserEmailVerified) error
}

type userEmailVerifiedUsecase struct {
	userRepository repository.UserRepository
	email          email.Email
	env            env.Env
	hash           hash.Hash
	log            log.Log
}

func NewUserEmailVerifiedUsecase(
	userRepository repository.UserRepository,
	email email.Email,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) UserEmailVerifiedUsecase {
	return &userEmailVerifiedUsecase{
		userRepository: userRepository,
		email:          email,
		env:            env,
		hash:           hash,
		log:            log,
	}
}

func (u *userEmailVerifiedUsecase) Exec(ctx context.Context, i *event.UserEmailVerified) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if user.IsEmailVerified() {
		u.log.Info("sending email verified", "email", user.Email)
		if err := u.email.Send(
			ctx,
			email.EmailVerifiedSubject,
			user.Email,
			&email.Payload{
				TemplateName: email.EmailVerifiedTemplate,
				Data: &email.EmailVerified{
					Name:     user.Name,
					LoginURL: u.env.Get("APP_URL"),
				},
			},
		); err != nil {
			return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email verified")
		}
		u.log.Info("verified email sent", "email", user.Email)
	}
	return nil
}

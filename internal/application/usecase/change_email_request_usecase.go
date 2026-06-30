package usecase

import (
	"context"
	"errors"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type ChangeEmailRequestUsecase interface {
	Exec(ctx context.Context, i *ChangeEmailRequestUsecaseInput) error
}

type changeEmailRequestUsecase struct {
	userRepository      repository.UserRepository
	verificationRequest VerificationRequestUsecase
	env                 env.Env
	log                 log.Log
}

type ChangeEmailRequestUsecaseInput struct {
	UserID int
	Email  string
}

func NewChangeEmailRequestUsecase(
	userRepository repository.UserRepository,
	verificationRequest VerificationRequestUsecase,
	env env.Env,
	log log.Log,
) ChangeEmailRequestUsecase {
	return &changeEmailRequestUsecase{
		userRepository:      userRepository,
		verificationRequest: verificationRequest,
		env:                 env,
		log:                 log,
	}
}

func (u *changeEmailRequestUsecase) Exec(ctx context.Context, i *ChangeEmailRequestUsecaseInput) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if user.Email == i.Email {
		return exception.Throw(exception.UNPROCESSABLE_ENTITY, "email is the same as the current one")
	}

	user, err = u.userRepository.FindByEmail(ctx, i.Email)

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	if user != nil {
		return exception.Throw(exception.CONFLICT, "email already exists")
	}

	_, err = u.verificationRequest.Exec(
		ctx,
		&VerificationRequestUsecaseInput{
			UserID:           i.UserID,
			VerificationType: entity.ChangeEmail,
			TokenType:        OTPToken,
			TokenLength:      6,
			ExpiredDuration:  u.env.Get("APP_CHANGE_EMAIL_EXPIRED"),
			Payload:          i.Email,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

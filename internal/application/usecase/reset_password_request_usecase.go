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

type ResetPasswordRequestUsecase interface {
	Exec(ctx context.Context, email string) error
}

type resetPasswordRequestUsecase struct {
	userRepository      repository.UserRepository
	verificationRequest VerificationRequestUsecase
	env                 env.Env
	log                 log.Log
}

func NewResetPasswordRequestUsecase(
	userRepository repository.UserRepository,
	verificationRequest VerificationRequestUsecase,
	env env.Env,
	log log.Log,
) ResetPasswordRequestUsecase {
	return &resetPasswordRequestUsecase{
		userRepository:      userRepository,
		verificationRequest: verificationRequest,
		env:                 env,
		log:                 log,
	}
}

func (u *resetPasswordRequestUsecase) Exec(ctx context.Context, email string) error {
	user, err := u.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	_, err = u.verificationRequest.Exec(
		ctx,
		&VerificationRequestUsecaseInput{
			UserID:           user.ID,
			VerificationType: entity.ResetPassword,
			TokenType:        OTPToken,
			TokenLength:      6,
			ExpiredDuration:  u.env.Get("APP_RESET_PASSWORD_EXPIRED"),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

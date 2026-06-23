package usecase

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/log"
)

type ChangeEmailRequestUsecase interface {
	Exec(ctx context.Context, i *ChangeEmailRequestUsecaseInput) error
}

type changeEmailRequestUsecase struct {
	verificationRequest VerificationRequestUsecase
	env                 env.Env
	log                 log.Log
}

type ChangeEmailRequestUsecaseInput struct {
	UserID int
	Email  string
}

func NewChangeEmailRequestUsecase(
	verificationRequest VerificationRequestUsecase,
	env env.Env,
	log log.Log,
) ChangeEmailRequestUsecase {
	return &changeEmailRequestUsecase{
		verificationRequest: verificationRequest,
		env:                 env,
		log:                 log,
	}
}

func (u *changeEmailRequestUsecase) Exec(ctx context.Context, i *ChangeEmailRequestUsecaseInput) error {
	_, err := u.verificationRequest.Exec(
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

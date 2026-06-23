package usecase

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/log"
)

type EmailVerificationRequestUsecase interface {
	Exec(ctx context.Context, userID int) error
}

type emailVerificationRequestUsecase struct {
	verificationRequest VerificationRequestUsecase
	env                 env.Env
	log                 log.Log
}

func NewEmailVerificationRequestUsecase(
	verificationRequest VerificationRequestUsecase,
	env env.Env,
	log log.Log,
) EmailVerificationRequestUsecase {
	return &emailVerificationRequestUsecase{
		verificationRequest: verificationRequest,
		env:                 env,
		log:                 log,
	}
}

func (u *emailVerificationRequestUsecase) Exec(ctx context.Context, userID int) error {
	_, err := u.verificationRequest.Exec(
		ctx,
		&VerificationRequestUsecaseInput{
			UserID:           userID,
			VerificationType: entity.EmailVerification,
			TokenType:        TextToken,
			TokenLength:      32,
			ExpiredDuration:  u.env.Get("APP_EMAIL_VERIFICATION_EXPIRED"),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

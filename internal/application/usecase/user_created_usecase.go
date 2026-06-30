package usecase

import (
	"context"

	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-common/log"
)

type UserCreatedUsecase interface {
	Exec(ctx context.Context, i *event.UserCreated) error
}

type userCreatedUsecase struct {
	emailVerificationRequest EmailVerificationRequestUsecase
	log                      log.Log
}

func NewUserCreatedUsecase(
	emailVerificationRequest EmailVerificationRequestUsecase,
	log log.Log,
) UserCreatedUsecase {
	return &userCreatedUsecase{
		emailVerificationRequest: emailVerificationRequest,
		log:                      log,
	}
}

func (u *userCreatedUsecase) Exec(ctx context.Context, i *event.UserCreated) error {

	if err := u.emailVerificationRequest.Exec(ctx, i.UserID); err != nil {
		return err
	}
	return nil
}

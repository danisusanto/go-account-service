package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/dani-susanto/go-common/hash"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type ResetPasswordUsecase interface {
	Exec(ctx context.Context, i *ResetPasswordUsecaseInput) error
}

type resetPasswordUsecase struct {
	userRepository      repository.UserRepository
	verificationUsecase VerificationUsedUsecase
	event               event.Event
	hash                hash.Hash
	log                 log.Log
}

type ResetPasswordUsecaseInput struct {
	Email       string
	Token       string
	NewPassword string
}

func NewResetPasswordUsecase(
	userRepository repository.UserRepository,
	verificationUsecase VerificationUsedUsecase,
	event event.Event,
	hash hash.Hash,
	log log.Log,
) ResetPasswordUsecase {
	return &resetPasswordUsecase{
		userRepository:      userRepository,
		verificationUsecase: verificationUsecase,
		event:               event,
		hash:                hash,
		log:                 log,
	}
}

func (u *resetPasswordUsecase) Exec(ctx context.Context, i *ResetPasswordUsecaseInput) error {
	user, err := u.userRepository.FindByEmail(ctx, i.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.verificationUsecase.Exec(
		ctx,
		&VerificationUsedInput{
			UserID:           user.ID,
			Token:            i.Token,
			VerificationType: entity.ResetPassword,
		},
	); err != nil {
		return err
	}

	password, err := u.hash.Hash(i.NewPassword)
	if err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.userRepository.SetPassword(ctx, user.ID, password); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.event.Publish(ctx, &event.Payload{
		EventName: event.UserResetPasswordEvent,
		CreatedAt: time.Now(),
		Data: &event.UserResetPassword{
			UserID: user.ID,
		},
	}); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user reset password event")
	}

	return nil
}

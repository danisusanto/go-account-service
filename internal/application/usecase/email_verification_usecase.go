package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type EmailVerificationUsecase interface {
	Exec(ctx context.Context, i *EmailVerificationUsecaseInput) error
}

type emailVerificationUsecase struct {
	userRepository      repository.UserRepository
	verificationUsecase VerificationVerifyUsecase
	event               event.Event
	log                 log.Log
}

type EmailVerificationUsecaseInput struct {
	Email string
	Token string
}

func NewEmailVerificationUsecase(
	userRepository repository.UserRepository,
	verificationUsecase VerificationVerifyUsecase,
	event event.Event,
	log log.Log,
) EmailVerificationUsecase {
	return &emailVerificationUsecase{
		userRepository:      userRepository,
		verificationUsecase: verificationUsecase,
		event:               event,
		log:                 log,
	}
}

func (u *emailVerificationUsecase) Exec(ctx context.Context, i *EmailVerificationUsecaseInput) error {
	user, err := u.userRepository.FindByEmail(ctx, i.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if user.IsEmailVerified() {
		return nil
	}

	if _, err := u.verificationUsecase.Exec(
		ctx,
		&VerificationVerifyUsecaseInput{
			UserID:           user.ID,
			Token:            i.Token,
			VerificationType: entity.EmailVerification,
		},
	); err != nil {
		return err
	}

	user.SetEmailVerified()
	if err := u.userRepository.MarkEmailAsVerified(ctx, user); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.event.Publish(ctx, &event.Payload{
		EventName: event.UserEmailVerifiedEvent,
		CreatedAt: time.Now(),
		Data: &event.UserEmailVerified{
			UserID: user.ID,
		},
	}); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user email verified event")
	}

	return nil
}

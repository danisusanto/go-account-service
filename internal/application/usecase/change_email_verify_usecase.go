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

type ChangeEmailVerifyUsecase interface {
	Exec(ctx context.Context, i *ChangeEmailVerifyUsecaseInput) error
}

type changeEmailVerifyUsecase struct {
	userRepository      repository.UserRepository
	verificationUsecase VerificationVerifyUsecase
	event               event.Event
	hash                hash.Hash
	log                 log.Log
}

type ChangeEmailVerifyUsecaseInput struct {
	UserID int
	Token  string
}

func NewChangeEmailVerifyUsecase(
	userRepository repository.UserRepository,
	verificationUsecase VerificationVerifyUsecase,
	event event.Event,
	hash hash.Hash,
	log log.Log,
) ChangeEmailVerifyUsecase {
	return &changeEmailVerifyUsecase{
		userRepository:      userRepository,
		verificationUsecase: verificationUsecase,
		event:               event,
		hash:                hash,
		log:                 log,
	}
}

func (u *changeEmailVerifyUsecase) Exec(ctx context.Context, i *ChangeEmailVerifyUsecaseInput) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	verification, err := u.verificationUsecase.Exec(
		ctx,
		&VerificationVerifyUsecaseInput{
			UserID:           user.ID,
			Token:            i.Token,
			VerificationType: entity.ChangeEmail,
		},
	)
	if err != nil {
		return err
	}

	if err := u.userRepository.SetEmail(ctx, user.ID, verification.Payload); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if err := u.event.Publish(ctx, &event.Payload{
		EventName: event.UserEmailChangedEvent,
		CreatedAt: time.Now(),
		Data: &event.UserEmailChanged{
			UserID: user.ID,
		},
	}); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user change email event")
	}

	return nil
}

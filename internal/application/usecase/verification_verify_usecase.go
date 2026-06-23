package usecase

import (
	"context"
	"errors"
	"strconv"

	"github.com/dani-susanto/go-common/hash"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/log"
)

type VerificationVerifyUsecase interface {
	Exec(ctx context.Context, i *VerificationVerifyUsecaseInput) (*VerificationVerifyUsecaseOutput, error)
}

type verificationVerifyUsecase struct {
	userVerificationRepository repository.UserVerificationRepository
	env                        env.Env
	hash                       hash.Hash
	log                        log.Log
}

type VerificationVerifyUsecaseInput struct {
	UserID           int
	Token            string
	VerificationType entity.VerificationType
}

type VerificationVerifyUsecaseOutput struct {
	ID      int
	Payload string
}

func NewVerificationVerifyUsecase(
	userVerificationRepository repository.UserVerificationRepository,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) VerificationVerifyUsecase {
	return &verificationVerifyUsecase{
		userVerificationRepository: userVerificationRepository,
		env:                        env,
		hash:                       hash,
		log:                        log,
	}
}

func (u *verificationVerifyUsecase) Exec(ctx context.Context, i *VerificationVerifyUsecaseInput) (*VerificationVerifyUsecaseOutput, error) {
	uev, err := u.userVerificationRepository.GetLatestActiveByUserID(ctx, i.UserID, i.VerificationType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.BAD_REQUEST, "verification not found or already expired")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	maxAttempt, err := strconv.Atoi(u.env.Get("APP_VERIFICATION_MAX_ATTEMPT"))
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	if uev.MaxAttemptExceeded(maxAttempt) {
		return nil, exception.Throw(exception.BAD_REQUEST, "verification max attempt exceeded")
	}

	if err := u.userVerificationRepository.IncrementAttempt(ctx, uev.ID, i.UserID, i.VerificationType); err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if !u.hash.Compare(i.Token, uev.TokenHash) {
		return nil, exception.Throw(exception.BAD_REQUEST, "verification token is invalid")
	}

	if uev.IsVerified() {
		return &VerificationVerifyUsecaseOutput{
			ID:      uev.ID,
			Payload: uev.Payload,
		}, nil
	}

	if err := u.userVerificationRepository.MarkAsVerified(ctx, uev); err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if uev.VerificationType == entity.EmailVerification || uev.VerificationType == entity.ChangeEmail {
		if err := u.userVerificationRepository.MarkAsUsed(ctx, uev); err != nil {
			return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
		}
	}

	return &VerificationVerifyUsecaseOutput{
		ID:      uev.ID,
		Payload: uev.Payload,
	}, nil
}

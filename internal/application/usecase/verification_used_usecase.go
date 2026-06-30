package usecase

import (
	"context"
	"errors"

	"github.com/danisusanto/go-common/hash"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type VerificationUsedUsecase interface {
	Exec(ctx context.Context, i *VerificationUsedInput) error
}

type verificationUsedUsecase struct {
	userVerificationRepository repository.UserVerificationRepository
	env                        env.Env
	hash                       hash.Hash
	log                        log.Log
}

type VerificationUsedInput struct {
	UserID           int
	Token            string
	VerificationType entity.VerificationType
}

func NewVerificationUsedUsecase(
	userVerificationRepository repository.UserVerificationRepository,
	env env.Env,
	hash hash.Hash,
	log log.Log,
) VerificationUsedUsecase {
	return &verificationUsedUsecase{
		userVerificationRepository: userVerificationRepository,
		env:                        env,
		hash:                       hash,
		log:                        log,
	}
}

func (u *verificationUsedUsecase) Exec(ctx context.Context, i *VerificationUsedInput) error {
	uev, err := u.userVerificationRepository.GetLatestActiveVerifiedByUserID(ctx, i.UserID, i.VerificationType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.BAD_REQUEST, "verification not found, already expired or used")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if !u.hash.Compare(i.Token, uev.TokenHash) {
		return exception.Throw(exception.BAD_REQUEST, "verification token is invalid")
	}

	if err := u.userVerificationRepository.MarkAsUsed(ctx, uev); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	return nil
}

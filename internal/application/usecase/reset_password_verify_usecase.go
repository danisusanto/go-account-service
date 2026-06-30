package usecase

import (
	"context"
	"errors"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type ResetPasswordVerifyUsecase interface {
	Exec(ctx context.Context, i *ResetPasswordVerifyUsecaseInput) error
}

type resetPasswordVerifyUsecase struct {
	userRepository      repository.UserRepository
	verificationUsecase VerificationVerifyUsecase
	log                 log.Log
}

type ResetPasswordVerifyUsecaseInput struct {
	Email string
	Token string
}

func NewResetPasswordVerifyUsecase(
	userRepository repository.UserRepository,
	verificationUsecase VerificationVerifyUsecase,
	log log.Log,
) ResetPasswordVerifyUsecase {
	return &resetPasswordVerifyUsecase{
		userRepository:      userRepository,
		verificationUsecase: verificationUsecase,
		log:                 log,
	}
}

func (u *resetPasswordVerifyUsecase) Exec(ctx context.Context, i *ResetPasswordVerifyUsecaseInput) error {
	user, err := u.userRepository.FindByEmail(ctx, i.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if _, err := u.verificationUsecase.Exec(
		ctx,
		&VerificationVerifyUsecaseInput{
			UserID:           user.ID,
			Token:            i.Token,
			VerificationType: entity.ResetPassword,
		},
	); err != nil {
		return err
	}

	return nil
}

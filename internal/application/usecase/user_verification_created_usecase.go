package usecase

import (
	"context"
	"errors"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type UserVerificationCreatedUsecase interface {
	Exec(ctx context.Context, i *event.UserVerificationCreated) error
}

type userVerificationCreatedUsecase struct {
	userRepository          repository.UserRepository
	emailVerificationNotify EmailVerificationNotifyUsecase
	resetPasswordNotify     ResetPasswordNotifyUsecase
	changeEmailNotify       ChangeEmailNotifyUsecase
	log                     log.Log
}

func NewUserVerificationCreatedUsecase(
	userRepository repository.UserRepository,
	emailVerificationNotify EmailVerificationNotifyUsecase,
	resetPasswordNotify ResetPasswordNotifyUsecase,
	changeEmailNotify ChangeEmailNotifyUsecase,
	log log.Log,
) UserVerificationCreatedUsecase {
	return &userVerificationCreatedUsecase{
		userRepository:          userRepository,
		emailVerificationNotify: emailVerificationNotify,
		resetPasswordNotify:     resetPasswordNotify,
		changeEmailNotify:       changeEmailNotify,
		log:                     log,
	}
}

func (u *userVerificationCreatedUsecase) Exec(ctx context.Context, i *event.UserVerificationCreated) error {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}
	switch i.VerificationType {
	case entity.EmailVerification:
		if err := u.emailVerificationNotify.Exec(
			ctx,
			&EmailVerificationNotifyUsecaseInput{
				Email: user.Email,
				Name:  user.Name,
				Token: i.PlainToken,
			},
		); err != nil {
			return err
		}
	case entity.ResetPassword:
		if err := u.resetPasswordNotify.Exec(
			ctx,
			&ResetPasswordNotifyUsecaseInput{
				Email: user.Email,
				Name:  user.Name,
				Token: i.PlainToken,
			},
		); err != nil {
			return err
		}
	case entity.ChangeEmail:
		if err := u.changeEmailNotify.Exec(
			ctx,
			&ChangeEmailNotifyUsecaseInput{
				Email: user.Email,
				Name:  user.Name,
				Token: i.PlainToken,
			},
		); err != nil {
			return err
		}
	default:
		if err := u.emailVerificationNotify.Exec(
			ctx,
			&EmailVerificationNotifyUsecaseInput{
				Email: user.Email,
				Name:  user.Name,
				Token: i.PlainToken,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

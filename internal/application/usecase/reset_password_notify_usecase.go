package usecase

import (
	"context"
	"strings"

	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type ResetPasswordNotifyUsecase interface {
	Exec(ctx context.Context, i *ResetPasswordNotifyUsecaseInput) error
}

type resetPasswordNotifyUsecase struct {
	email email.Email
	env   env.Env
	log   log.Log
}

type ResetPasswordNotifyUsecaseInput struct {
	Email string
	Name  string
	Token string
}

func NewResetPasswordNotifyUsecase(
	email email.Email,
	env env.Env,
	log log.Log,
) ResetPasswordNotifyUsecase {
	return &resetPasswordNotifyUsecase{
		email: email,
		env:   env,
		log:   log,
	}
}

func (u *resetPasswordNotifyUsecase) Exec(ctx context.Context, i *ResetPasswordNotifyUsecaseInput) error {
	u.log.Info("sending reset password email", "email", i.Email)
	if err := u.email.Send(
		ctx,
		email.ResetPasswordSubject,
		i.Email,
		&email.Payload{
			TemplateName: email.ResetPasswordTemplate,
			Data: &email.ResetPassword{
				Name:      i.Name,
				Token:     i.Token,
				ExpiredIn: strings.ReplaceAll(u.env.Get("APP_RESET_PASSWORD_EXPIRED"), "m", ""),
			},
		},
	); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send reset password email")
	}
	u.log.Info("reset password email sent", "email", i.Email)

	return nil
}

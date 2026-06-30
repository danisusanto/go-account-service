package usecase

import (
	"context"
	"strings"

	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type ChangeEmailNotifyUsecase interface {
	Exec(ctx context.Context, i *ChangeEmailNotifyUsecaseInput) error
}

type changeEmailNotifyUsecase struct {
	email email.Email
	env   env.Env
	log   log.Log
}

type ChangeEmailNotifyUsecaseInput struct {
	Email string
	Name  string
	Token string
}

func NewChangeEmailNotifyUsecase(
	email email.Email,
	env env.Env,
	log log.Log,
) ChangeEmailNotifyUsecase {
	return &changeEmailNotifyUsecase{
		email: email,
		env:   env,
		log:   log,
	}
}

func (u *changeEmailNotifyUsecase) Exec(ctx context.Context, i *ChangeEmailNotifyUsecaseInput) error {
	u.log.Info("sending change email token", "email", i.Email)
	if err := u.email.Send(
		ctx,
		email.ChangeEmailSubject,
		i.Email,
		&email.Payload{
			TemplateName: email.ChangeEmailTemplate,
			Data: &email.ChangeEmail{
				Name:      i.Name,
				Token:     i.Token,
				ExpiredIn: strings.ReplaceAll(u.env.Get("APP_CHANGE_EMAIL_EXPIRED"), "m", ""),
			},
		},
	); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send change email token")
	}
	u.log.Info("change email token sent", "email", i.Email)

	return nil
}

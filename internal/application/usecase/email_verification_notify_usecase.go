package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-common/env"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type EmailVerificationNotifyUsecase interface {
	Exec(ctx context.Context, i *EmailVerificationNotifyUsecaseInput) error
}

type emailVerificationNotifyUsecase struct {
	email email.Email
	env   env.Env
	log   log.Log
}

type EmailVerificationNotifyUsecaseInput struct {
	Email string
	Name  string
	Token string
}

func NewEmailVerificationNotifyUsecase(
	email email.Email,
	env env.Env,
	log log.Log,
) EmailVerificationNotifyUsecase {
	return &emailVerificationNotifyUsecase{
		email: email,
		env:   env,
		log:   log,
	}
}

func (u *emailVerificationNotifyUsecase) Exec(ctx context.Context, i *EmailVerificationNotifyUsecaseInput) error {
	u.log.Info("sending email verification", "email", i.Email)
	if err := u.email.Send(
		ctx,
		email.EmailVerificationSubject,
		i.Email,
		&email.Payload{
			TemplateName: email.EmailVerificationTemplate,
			Data: &email.EmailVerification{
				Name: i.Name,
				VerificationURL: fmt.Sprintf(
					"%s?source=%s&verify=%s",
					u.env.Get("APP_EMAIL_VERIFICATION_URL"),
					i.Email,
					i.Token,
				),
				ExpiredIn: strings.ReplaceAll(u.env.Get("APP_EMAIL_VERIFICATION_EXPIRED"), "m", ""),
			},
		},
	); err != nil {
		return exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email verification")
	}
	u.log.Info("verification email sent", "email", i.Email)

	return nil
}

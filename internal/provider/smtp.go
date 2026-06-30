package provider

import (
	"strconv"

	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-account-service/internal/infrastructure/notification/email/smtp"
	"github.com/danisusanto/go-common/env"
	commonsmtp "github.com/danisusanto/go-common/smtp"
)

func SetupSMTP(env env.Env) (email.Email, error) {
	port, err := strconv.Atoi(env.Get("SMTP_PORT"))
	if err != nil {
		return nil, err
	}

	tls, err := strconv.ParseBool(env.Get("SMTP_TLS"))
	if err != nil {
		return nil, err
	}

	gosmtp := commonsmtp.New(
		env.Get("SMTP_HOST"),
		port,
		env.Get("SMTP_USERNAME"),
		env.Get("SMTP_PASSWORD"),
		env.Get("SMTP_FROM"),
		tls,
	)

	return smtp.NewSMTP(gosmtp, env), nil
}

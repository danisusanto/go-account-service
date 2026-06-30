package smtp

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"time"

	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-common/env"
	commonsmtp "github.com/danisusanto/go-common/smtp"
)

type smtp struct {
	gosmtp commonsmtp.SMTP
	env    env.Env
}

func NewSMTP(
	gosmtp commonsmtp.SMTP,
	env env.Env,
) email.Email {
	return &smtp{
		gosmtp: gosmtp,
		env:    env,
	}
}

func (m *smtp) Send(ctx context.Context, subject string, to string, payload *email.Payload) error {
	domainBytes, err := json.Marshal(payload.Data)
	if err != nil {
		return err
	}

	result := map[string]any{
		"app_name":        m.env.Get("APP_NAME"),
		"app_address":     m.env.Get("APP_ADDRESS"),
		"year":            time.Now().Year(),
		"unsubscribe_url": m.env.Get("APP_UNSUBSCRIBE_URL"),
		"privacy_url":     m.env.Get("APP_PRIVACY_URL"),
		"subject":         subject,
	}

	if err := json.Unmarshal(domainBytes, &result); err != nil {
		return err
	}

	content, err := m.render(payload.TemplateName, result)
	if err != nil {
		return err
	}

	if err := m.gosmtp.Send(subject, to, content); err != nil {
		return err
	}
	return nil
}

func (t *smtp) render(templateName string, data any) (string, error) {
	tmpl, err := template.ParseFiles(
		"templates/email/layout.html",
		"templates/email/"+templateName,
	)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

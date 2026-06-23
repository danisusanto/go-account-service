package email

import "context"

type Payload struct {
	TemplateName string `json:"template_name"`
	TemplateID   string `json:"template_id"`
	Data         any    `json:"data"`
}

type Email interface {
	Send(ctx context.Context, subject string, to string, payload *Payload) error
}

package event

import "github.com/danisusanto/go-account-service/internal/domain/entity"

const UserVerificationCreatedEvent = "account.user.verification.created"

type UserVerificationCreated struct {
	UserID           int
	VerificationID   int
	VerificationType entity.VerificationType
	PlainToken       string
}

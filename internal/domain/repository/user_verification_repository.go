package repository

import (
	"context"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
)

type UserVerificationRepository interface {
	Create(ctx context.Context, uev *entity.UserVerification) (*entity.UserVerification, error)
	GetLatestActiveByUserID(ctx context.Context, userID int, verificationType entity.VerificationType) (*entity.UserVerification, error)
	GetLatestActiveVerifiedByUserID(ctx context.Context, userID int, verificationType entity.VerificationType) (*entity.UserVerification, error)
	IncrementAttempt(ctx context.Context, id int, userID int, verificationType entity.VerificationType) error
	MarkAsVerified(ctx context.Context, uev *entity.UserVerification) error
	MarkAsUsed(ctx context.Context, uev *entity.UserVerification) error
}

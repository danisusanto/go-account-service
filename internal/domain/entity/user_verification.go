package entity

import "time"

type VerificationType string

const (
	EmailVerification VerificationType = "EMAIL_VERIFICATION"
	ResetPassword     VerificationType = "RESET_PASSWORD"
	ChangeEmail       VerificationType = "CHANGE_EMAIL"
)

type UserVerification struct {
	ID               int
	UserID           int
	VerificationType VerificationType
	TokenHash        string
	Attempt          int
	Payload          string
	CreatedBy        int
	UpdatedBy        int
	DeletedBy        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	VerifiedAt       *time.Time
	UsedAt           *time.Time
	ExpiredAt        time.Time
	DeletedAt        time.Time
}

func NewUserVerification(
	userID int,
	verificationType VerificationType,
	tokenHash string,
	payload string,
	expiredAt time.Time,
) *UserVerification {
	return &UserVerification{
		UserID:           userID,
		VerificationType: verificationType,
		TokenHash:        tokenHash,
		Attempt:          0,
		Payload:          payload,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ExpiredAt:        expiredAt,
	}
}

func (e *UserVerification) MaxAttemptExceeded(maxAttempt int) bool {
	return e.Attempt >= maxAttempt
}

func (e *UserVerification) IsVerified() bool {
	return e.VerifiedAt != nil
}

func (e *UserVerification) IsUsed() bool {
	return e.UsedAt != nil
}

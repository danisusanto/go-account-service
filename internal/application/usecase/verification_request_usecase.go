package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/danisusanto/go-common/hash"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/log"
)

type VerificationRequestUsecase interface {
	Exec(ctx context.Context, i *VerificationRequestUsecaseInput) (*VerificationRequestUsecaseOutput, error)
}

type verificationRequestUsecase struct {
	userRepository             repository.UserRepository
	userVerificationRepository repository.UserVerificationRepository
	event                      event.Event
	hash                       hash.Hash
	log                        log.Log
}

type TokenType string

const (
	TextToken TokenType = "TEXT"
	OTPToken  TokenType = "OTP"
)

type VerificationRequestUsecaseInput struct {
	UserID           int
	VerificationType entity.VerificationType
	TokenType        TokenType
	TokenLength      int
	ExpiredDuration  string
	Payload          string
}

type VerificationRequestUsecaseOutput struct {
	PlainToken   string
	User         *entity.User
	Verification *entity.UserVerification
}

func NewVerificationRequestUsecase(
	userRepository repository.UserRepository,
	userVerificationRepository repository.UserVerificationRepository,
	event event.Event,
	hash hash.Hash,
	log log.Log,
) VerificationRequestUsecase {
	return &verificationRequestUsecase{
		userRepository:             userRepository,
		userVerificationRepository: userVerificationRepository,
		event:                      event,
		hash:                       hash,
		log:                        log,
	}
}

func (u *verificationRequestUsecase) Exec(ctx context.Context, i *VerificationRequestUsecaseInput) (*VerificationRequestUsecaseOutput, error) {
	user, err := u.userRepository.Get(ctx, i.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, exception.Throw(exception.NOT_FOUND, "user not found")
		}
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, err.Error())
	}

	if i.VerificationType == entity.EmailVerification && user.IsEmailVerified() {
		return nil, exception.Throw(exception.BAD_REQUEST, "email already verified")
	}

	var token string
	switch i.TokenType {
	case OTPToken:
		token, err = u.hash.GenerateOTP(i.TokenLength)
	default:
		token, err = u.hash.GenerateToken(i.TokenLength)
	}
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate random token")
	}
	tokenHash, err := u.hash.Hash(token)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash verification token")
	}

	tokenVerificationExpiredDuration, err := time.ParseDuration(i.ExpiredDuration)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed parse verification expired duration")
	}
	verification, err := u.userVerificationRepository.Create(
		ctx, entity.NewUserVerification(
			user.ID,
			i.VerificationType,
			tokenHash,
			i.Payload,
			time.Now().Add(tokenVerificationExpiredDuration),
		),
	)
	if err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create user verification")
	}

	if err := u.event.Publish(ctx, &event.Payload{
		EventName: event.UserVerificationCreatedEvent,
		CreatedAt: time.Now(),
		Data: &event.UserVerificationCreated{
			UserID:           user.ID,
			VerificationID:   verification.ID,
			VerificationType: verification.VerificationType,
			PlainToken:       token,
		},
	}); err != nil {
		return nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user verification created event")
	}

	return &VerificationRequestUsecaseOutput{
		PlainToken:   token,
		User:         user,
		Verification: verification,
	}, nil
}

package usecase_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mockenv "github.com/dani-susanto/go-account-service/mocks/env"
	mockhash "github.com/dani-susanto/go-account-service/mocks/hash"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
)

type mockTestVerificationVerifyUsecase struct {
	userVerificationRepository *mockrepository.MockUserVerificationRepository
	env                        *mockenv.MockEnv
	hash                       *mockhash.MockHash
	log                        *mocklog.MockLog
}

func TestVerificationVerifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.VerificationVerifyUsecaseInput{UserID: 1, Token: "token", VerificationType: entity.EmailVerification}

	tests := []struct {
		name      string
		input     *usecase.VerificationVerifyUsecaseInput
		output    *usecase.VerificationVerifyUsecaseOutput
		err       error
		setupMock func(m *mockTestVerificationVerifyUsecase)
	}{
		{
			name:  "verification not found",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification not found or already expired"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "invalid max attempt config",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "strconv.Atoi: parsing \"invalid\": invalid syntax"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("invalid")
			},
		},
		{
			name:  "max attempt exceeded",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification max attempt exceeded"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				verification.Attempt = 3
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
			},
		},
		{
			name:  "failed to increment attempt",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to increment attempt"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to increment attempt"))
			},
		},
		{
			name:  "invalid token",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(false)
			},
		},
		{
			name:  "already verified",
			input: input,
			output: &usecase.VerificationVerifyUsecaseOutput{
				ID:      1,
				Payload: "payload",
			},
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				now := time.Now()
				verification.VerifiedAt = &now
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
			},
		},
		{
			name:  "failed to mark as verified",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as verified"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsVerified", ctx, verification).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as verified"))
			},
		},
		{
			name:  "failed to mark as used",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as used"),
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsVerified", ctx, verification).Return(nil)
				m.userVerificationRepository.On("MarkAsUsed", ctx, verification).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as used"))
			},
		},
		{
			name:  "success",
			input: input,
			output: &usecase.VerificationVerifyUsecaseOutput{
				ID:      1,
				Payload: "payload",
			},
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				verification := newTestVerification(input.UserID, input.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return("3")
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, input.UserID, input.VerificationType).Return(nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsVerified", ctx, verification).Return(nil)
				m.userVerificationRepository.On("MarkAsUsed", ctx, verification).Return(nil)
			},
		},
		{
			name:  "success without mark as used",
			input: &usecase.VerificationVerifyUsecaseInput{UserID: 1, Token: "token", VerificationType: entity.ResetPassword},
			output: &usecase.VerificationVerifyUsecaseOutput{
				ID:      1,
				Payload: "payload",
			},
			setupMock: func(m *mockTestVerificationVerifyUsecase) {
				i := &usecase.VerificationVerifyUsecaseInput{UserID: 1, Token: "token", VerificationType: entity.ResetPassword}
				verification := newTestVerification(i.UserID, i.VerificationType)
				m.userVerificationRepository.On("GetLatestActiveByUserID", ctx, i.UserID, i.VerificationType).Return(verification, nil)
				m.env.On("Get", "APP_VERIFICATION_MAX_ATTEMPT").Return(strconv.Itoa(3))
				m.userVerificationRepository.On("IncrementAttempt", ctx, verification.ID, i.UserID, i.VerificationType).Return(nil)
				m.hash.On("Compare", i.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsVerified", ctx, verification).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestVerificationVerifyUsecase{
				userVerificationRepository: mockrepository.NewMockUserVerificationRepository(t),
				env:                        mockenv.NewMockEnv(t),
				hash:                       mockhash.NewMockHash(t),
				log:                        mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewVerificationVerifyUsecase(mock.userVerificationRepository, mock.env, mock.hash, mock.log)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, tt.output, res)
		})
	}
}

func newTestVerification(userID int, verificationType entity.VerificationType) *entity.UserVerification {
	verification := entity.NewUserVerification(userID, verificationType, "tokenhash", "payload", time.Now().Add(time.Minute))
	verification.ID = 1
	return verification
}

func timeNow() time.Time {
	return time.Now().Add(time.Minute)
}

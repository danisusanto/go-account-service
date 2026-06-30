package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockenv "github.com/danisusanto/go-account-service/mocks/env"
	mockhash "github.com/danisusanto/go-account-service/mocks/hash"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
)

type mockTestVerificationUsedUsecase struct {
	userVerificationRepository *mockrepository.MockUserVerificationRepository
	env                        *mockenv.MockEnv
	hash                       *mockhash.MockHash
	log                        *mocklog.MockLog
}

func TestVerificationUsedUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.VerificationUsedInput{UserID: 1, Token: "token", VerificationType: entity.ResetPassword}
	verification := entity.NewUserVerification(input.UserID, input.VerificationType, "tokenhash", "", timeNow())
	verification.ID = 1

	tests := []struct {
		name      string
		input     *usecase.VerificationUsedInput
		err       error
		setupMock func(m *mockTestVerificationUsedUsecase)
	}{
		{
			name:  "verification not found",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification not found, already expired or used"),
			setupMock: func(m *mockTestVerificationUsedUsecase) {
				m.userVerificationRepository.On("GetLatestActiveVerifiedByUserID", ctx, input.UserID, input.VerificationType).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestVerificationUsedUsecase) {
				m.userVerificationRepository.On("GetLatestActiveVerifiedByUserID", ctx, input.UserID, input.VerificationType).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "invalid token",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestVerificationUsedUsecase) {
				m.userVerificationRepository.On("GetLatestActiveVerifiedByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(false)
			},
		},
		{
			name:  "failed to mark as used",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as used"),
			setupMock: func(m *mockTestVerificationUsedUsecase) {
				m.userVerificationRepository.On("GetLatestActiveVerifiedByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsUsed", ctx, verification).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark as used"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestVerificationUsedUsecase) {
				m.userVerificationRepository.On("GetLatestActiveVerifiedByUserID", ctx, input.UserID, input.VerificationType).Return(verification, nil)
				m.hash.On("Compare", input.Token, verification.TokenHash).Return(true)
				m.userVerificationRepository.On("MarkAsUsed", ctx, verification).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestVerificationUsedUsecase{
				userVerificationRepository: mockrepository.NewMockUserVerificationRepository(t),
				env:                        mockenv.NewMockEnv(t),
				hash:                       mockhash.NewMockHash(t),
				log:                        mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewVerificationUsedUsecase(mock.userVerificationRepository, mock.env, mock.hash, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

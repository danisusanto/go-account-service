package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	mockenv "github.com/danisusanto/go-account-service/mocks/env"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestEmailVerificationRequestUsecase struct {
	verificationRequest *mockusecase.MockVerificationRequestUsecase
	env                 *mockenv.MockEnv
	log                 *mocklog.MockLog
}

func TestEmailVerificationRequestUsecase(t *testing.T) {
	ctx := context.Background()
	userID := 1

	tests := []struct {
		name      string
		userID    int
		err       error
		setupMock func(m *mockTestEmailVerificationRequestUsecase)
	}{
		{
			name:   "verification request failed",
			userID: userID,
			err:    exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"),
			setupMock: func(m *mockTestEmailVerificationRequestUsecase) {
				m.env.On("Get", "APP_EMAIL_VERIFICATION_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == userID &&
						i.VerificationType == entity.EmailVerification &&
						i.TokenType == usecase.TextToken &&
						i.TokenLength == 32 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == ""
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"))
			},
		},
		{
			name:   "success",
			userID: userID,
			setupMock: func(m *mockTestEmailVerificationRequestUsecase) {
				m.env.On("Get", "APP_EMAIL_VERIFICATION_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == userID &&
						i.VerificationType == entity.EmailVerification &&
						i.TokenType == usecase.TextToken &&
						i.TokenLength == 32 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == ""
				})).Return(&usecase.VerificationRequestUsecaseOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestEmailVerificationRequestUsecase{
				verificationRequest: mockusecase.NewMockVerificationRequestUsecase(t),
				env:                 mockenv.NewMockEnv(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewEmailVerificationRequestUsecase(mock.verificationRequest, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.userID)

			assert.Equal(t, tt.err, err)
		})
	}
}

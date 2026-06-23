package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mockenv "github.com/dani-susanto/go-account-service/mocks/env"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestResetPasswordRequestUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationRequest *mockusecase.MockVerificationRequestUsecase
	env                 *mockenv.MockEnv
	log                 *mocklog.MockLog
}

func TestResetPasswordRequestUsecase(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	user := entity.NewUser("test", email, "hashedpassword", "62", "62812345678")
	user.ID = 1

	tests := []struct {
		name      string
		email     string
		err       error
		setupMock func(m *mockTestResetPasswordRequestUsecase)
	}{
		{
			name:  "user not found",
			email: email,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestResetPasswordRequestUsecase) {
				m.userRepository.On("FindByEmail", ctx, email).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			email: email,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestResetPasswordRequestUsecase) {
				m.userRepository.On("FindByEmail", ctx, email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "verification request failed",
			email: email,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"),
			setupMock: func(m *mockTestResetPasswordRequestUsecase) {
				m.userRepository.On("FindByEmail", ctx, email).Return(user, nil)
				m.env.On("Get", "APP_RESET_PASSWORD_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.VerificationType == entity.ResetPassword &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == ""
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"))
			},
		},
		{
			name:  "success",
			email: email,
			setupMock: func(m *mockTestResetPasswordRequestUsecase) {
				m.userRepository.On("FindByEmail", ctx, email).Return(user, nil)
				m.env.On("Get", "APP_RESET_PASSWORD_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.VerificationType == entity.ResetPassword &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == ""
				})).Return(&usecase.VerificationRequestUsecaseOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordRequestUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationRequest: mockusecase.NewMockVerificationRequestUsecase(t),
				env:                 mockenv.NewMockEnv(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewResetPasswordRequestUsecase(
				mock.userRepository,
				mock.verificationRequest,
				mock.env,
				mock.log,
			)

			err := usecase.Exec(ctx, tt.email)

			assert.Equal(t, tt.err, err)
		})
	}
}

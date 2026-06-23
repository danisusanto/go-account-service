package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestResetPasswordVerifyUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationUsecase *mockusecase.MockVerificationVerifyUsecase
	log                 *mocklog.MockLog
}

func TestResetPasswordVerifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ResetPasswordVerifyUsecaseInput{Email: "test@example.com", Token: "token"}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")

	tests := []struct {
		name      string
		input     *usecase.ResetPasswordVerifyUsecaseInput
		err       error
		setupMock func(m *mockTestResetPasswordVerifyUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestResetPasswordVerifyUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestResetPasswordVerifyUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "verification failed",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestResetPasswordVerifyUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ResetPassword
				})).Return(nil, exception.Throw(exception.BAD_REQUEST, "verification token is invalid"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestResetPasswordVerifyUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ResetPassword
				})).Return(&usecase.VerificationVerifyUsecaseOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordVerifyUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationUsecase: mockusecase.NewMockVerificationVerifyUsecase(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewResetPasswordVerifyUsecase(mock.userRepository, mock.verificationUsecase, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

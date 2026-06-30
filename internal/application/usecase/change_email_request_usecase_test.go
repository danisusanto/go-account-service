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

type mockTestChangeEmailRequestUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationRequest *mockusecase.MockVerificationRequestUsecase
	env                 *mockenv.MockEnv
	log                 *mocklog.MockLog
}

func TestChangeEmailRequestUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ChangeEmailRequestUsecaseInput{
		UserID: 1,
		Email:  "test@example.com",
	}

	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = input.UserID
	currentUser := entity.NewUser("test", "old@example.com", "hashedpassword", "62", "62812345678")
	currentUser.ID = input.UserID

	tests := []struct {
		name      string
		input     *usecase.ChangeEmailRequestUsecaseInput
		err       error
		setupMock func(m *mockTestChangeEmailRequestUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email is the same as the current one",
			input: input,
			err:   exception.Throw(exception.UNPROCESSABLE_ENTITY, "email is the same as the current one"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
			},
		},
		{
			name:  "failed to find email",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(currentUser, nil)
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email already exists",
			input: input,
			err:   exception.Throw(exception.CONFLICT, "email already exists"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(currentUser, nil)
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
			},
		},
		{
			name:  "verification request failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(currentUser, nil)
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.env.On("Get", "APP_CHANGE_EMAIL_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == input.UserID &&
						i.VerificationType == entity.ChangeEmail &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == input.Email
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(currentUser, nil)
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.env.On("Get", "APP_CHANGE_EMAIL_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == input.UserID &&
						i.VerificationType == entity.ChangeEmail &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == input.Email
				})).Return(&usecase.VerificationRequestUsecaseOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailRequestUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationRequest: mockusecase.NewMockVerificationRequestUsecase(t),
				env:                 mockenv.NewMockEnv(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewChangeEmailRequestUsecase(mock.userRepository, mock.verificationRequest, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

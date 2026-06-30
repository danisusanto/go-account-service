package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockemail "github.com/danisusanto/go-account-service/mocks/email"
	mockenv "github.com/danisusanto/go-account-service/mocks/env"
	mockhash "github.com/danisusanto/go-account-service/mocks/hash"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUserResetPasswordUsecase struct {
	userRepository *mockrepository.MockUserRepository
	email          *mockemail.MockEmail
	env            *mockenv.MockEnv
	hash           *mockhash.MockHash
	log            *mocklog.MockLog
}

func TestUserResetPasswordUsecase(t *testing.T) {
	ctx := context.Background()
	input := &event.UserResetPassword{UserID: 1}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = input.UserID

	tests := []struct {
		name      string
		input     *event.UserResetPassword
		err       error
		setupMock func(m *mockTestUserResetPasswordUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestUserResetPasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestUserResetPasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "send email failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send reset password success email"),
			setupMock: func(m *mockTestUserResetPasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_URL").Return("https://example.com")
				m.email.On("Send", ctx, email.ResetPasswordSuccessSubject, user.Email, mock.MatchedBy(func(p *email.Payload) bool {
					data, ok := p.Data.(*email.ResetPasswordSuccess)
					return p.TemplateName == email.ResetPasswordSuccessTemplate &&
						ok &&
						data.Name == user.Name &&
						data.LoginURL == "https://example.com"
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestUserResetPasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_URL").Return("https://example.com")
				m.email.On("Send", ctx, email.ResetPasswordSuccessSubject, user.Email, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserResetPasswordUsecase{
				userRepository: mockrepository.NewMockUserRepository(t),
				email:          mockemail.NewMockEmail(t),
				env:            mockenv.NewMockEnv(t),
				hash:           mockhash.NewMockHash(t),
				log:            mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewUserResetPasswordUsecase(mock.userRepository, mock.email, mock.env, mock.hash, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mockevent "github.com/dani-susanto/go-account-service/mocks/event"
	mockhash "github.com/dani-susanto/go-account-service/mocks/hash"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestResetPasswordUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationUsecase *mockusecase.MockVerificationUsedUsecase
	event               *mockevent.MockEvent
	hash                *mockhash.MockHash
	log                 *mocklog.MockLog
}

func TestResetPasswordUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ResetPasswordUsecaseInput{Email: "test@example.com", Token: "token", NewPassword: "New123+"}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	hashedPassword := "newhashedpassword"

	tests := []struct {
		name      string
		input     *usecase.ResetPasswordUsecaseInput
		err       error
		setupMock func(m *mockTestResetPasswordUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "verification failed",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationUsedInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ResetPassword
				})).Return(exception.Throw(exception.BAD_REQUEST, "verification token is invalid"))
			},
		},
		{
			name:  "failed to hash password",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash password"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.Anything).Return(nil)
				m.hash.On("Hash", input.NewPassword).Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash password"))
			},
		},
		{
			name:  "failed to set password",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set password"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.Anything).Return(nil)
				m.hash.On("Hash", input.NewPassword).Return(hashedPassword, nil)
				m.userRepository.On("SetPassword", ctx, user.ID, hashedPassword).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set password"))
			},
		},
		{
			name:  "failed to publish user reset password event",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user reset password event"),
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.Anything).Return(nil)
				m.hash.On("Hash", input.NewPassword).Return(hashedPassword, nil)
				m.userRepository.On("SetPassword", ctx, user.ID, hashedPassword).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserResetPassword)
					return p.EventName == event.UserResetPasswordEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish event"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestResetPasswordUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.Anything).Return(nil)
				m.hash.On("Hash", input.NewPassword).Return(hashedPassword, nil)
				m.userRepository.On("SetPassword", ctx, user.ID, hashedPassword).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserResetPassword)
					return p.EventName == event.UserResetPasswordEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationUsecase: mockusecase.NewMockVerificationUsedUsecase(t),
				event:               mockevent.NewMockEvent(t),
				hash:                mockhash.NewMockHash(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewResetPasswordUsecase(mock.userRepository, mock.verificationUsecase, mock.event, mock.hash, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

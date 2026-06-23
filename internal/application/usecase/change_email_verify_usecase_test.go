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

type mockTestChangeEmailVerifyUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationUsecase *mockusecase.MockVerificationVerifyUsecase
	event               *mockevent.MockEvent
	hash                *mockhash.MockHash
	log                 *mocklog.MockLog
}

func TestChangeEmailVerifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ChangeEmailVerifyUsecaseInput{
		UserID: 1,
		Token:  "token",
	}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = input.UserID
	output := &usecase.VerificationVerifyUsecaseOutput{
		ID:      1,
		Payload: "new@example.com",
	}

	tests := []struct {
		name      string
		input     *usecase.ChangeEmailVerifyUsecaseInput
		err       error
		setupMock func(m *mockTestChangeEmailVerifyUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "verification failed",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ChangeEmail
				})).Return(nil, exception.Throw(exception.BAD_REQUEST, "verification token is invalid"))
			},
		},
		{
			name:  "failed to set email",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set email"),
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ChangeEmail
				})).Return(output, nil)
				m.userRepository.On("SetEmail", ctx, user.ID, output.Payload).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set email"))
			},
		},
		{
			name:  "failed to publish user change email event",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user change email event"),
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ChangeEmail
				})).Return(output, nil)
				m.userRepository.On("SetEmail", ctx, user.ID, output.Payload).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserEmailChanged)
					return p.EventName == event.UserEmailChangedEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish event"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestChangeEmailVerifyUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.ChangeEmail
				})).Return(output, nil)
				m.userRepository.On("SetEmail", ctx, user.ID, output.Payload).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserEmailChanged)
					return p.EventName == event.UserEmailChangedEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailVerifyUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationUsecase: mockusecase.NewMockVerificationVerifyUsecase(t),
				event:               mockevent.NewMockEvent(t),
				hash:                mockhash.NewMockHash(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewChangeEmailVerifyUsecase(
				mock.userRepository,
				mock.verificationUsecase,
				mock.event,
				mock.hash,
				mock.log,
			)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

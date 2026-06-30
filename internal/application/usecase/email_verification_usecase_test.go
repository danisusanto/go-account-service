package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockevent "github.com/danisusanto/go-account-service/mocks/event"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestEmailVerificationUsecase struct {
	userRepository      *mockrepository.MockUserRepository
	verificationUsecase *mockusecase.MockVerificationVerifyUsecase
	event               *mockevent.MockEvent
}

func TestEmailVerificationUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.EmailVerificationUsecaseInput{
		Email: "test@example.com",
		Token: "token",
	}

	tests := []struct {
		name      string
		input     *usecase.EmailVerificationUsecaseInput
		err       error
		setupMock func(m *mockTestEmailVerificationUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email already verified",
			input: input,
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				user := entity.NewUser("test", input.Email, "hashedpassword", "62", "62812345678")
				user.ID = 1
				user.SetEmailVerified()
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
			},
		},
		{
			name:  "verification failed",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "verification token is invalid"),
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				user := entity.NewUser("test", input.Email, "hashedpassword", "62", "62812345678")
				user.ID = 1
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.EmailVerification
				})).Return(nil, exception.Throw(exception.BAD_REQUEST, "verification token is invalid"))
			},
		},
		{
			name:  "failed to mark email as verified",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark email as verified"),
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				user := entity.NewUser("test", input.Email, "hashedpassword", "62", "62812345678")
				user.ID = 1
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.EmailVerification
				})).Return(&usecase.VerificationVerifyUsecaseOutput{ID: 1}, nil)
				m.userRepository.On("MarkEmailAsVerified", ctx, mock.MatchedBy(func(u *entity.User) bool {
					return u.ID == user.ID && u.IsEmailVerified()
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to mark email as verified"))
			},
		},
		{
			name:  "failed to publish user email verified event",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user email verified event"),
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				user := entity.NewUser("test", input.Email, "hashedpassword", "62", "62812345678")
				user.ID = 1
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.EmailVerification
				})).Return(&usecase.VerificationVerifyUsecaseOutput{ID: 1}, nil)
				m.userRepository.On("MarkEmailAsVerified", ctx, mock.MatchedBy(func(u *entity.User) bool {
					return u.ID == user.ID && u.IsEmailVerified()
				})).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserEmailVerified)
					return p.EventName == event.UserEmailVerifiedEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish event"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestEmailVerificationUsecase) {
				user := entity.NewUser("test", input.Email, "hashedpassword", "62", "62812345678")
				user.ID = 1
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(user, nil)
				m.verificationUsecase.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationVerifyUsecaseInput) bool {
					return i.UserID == user.ID &&
						i.Token == input.Token &&
						i.VerificationType == entity.EmailVerification
				})).Return(&usecase.VerificationVerifyUsecaseOutput{ID: 1}, nil)
				m.userRepository.On("MarkEmailAsVerified", ctx, mock.MatchedBy(func(u *entity.User) bool {
					return u.ID == user.ID && u.IsEmailVerified()
				})).Return(nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserEmailVerified)
					return p.EventName == event.UserEmailVerifiedEvent &&
						ok &&
						data.UserID == user.ID
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestEmailVerificationUsecase{
				userRepository:      mockrepository.NewMockUserRepository(t),
				verificationUsecase: mockusecase.NewMockVerificationVerifyUsecase(t),
				event:               mockevent.NewMockEvent(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewEmailVerificationUsecase(
				mock.userRepository,
				mock.verificationUsecase,
				mock.event,
				nil,
			)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

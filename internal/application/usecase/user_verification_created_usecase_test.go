package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUserVerificationCreatedUsecase struct {
	userRepository          *mockrepository.MockUserRepository
	emailVerificationNotify *mockusecase.MockEmailVerificationNotifyUsecase
	resetPasswordNotify     *mockusecase.MockResetPasswordNotifyUsecase
	changeEmailNotify       *mockusecase.MockChangeEmailNotifyUsecase
	log                     *mocklog.MockLog
}

func TestUserVerificationCreatedUsecase(t *testing.T) {
	ctx := context.Background()
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = 1
	input := &event.UserVerificationCreated{UserID: user.ID, VerificationID: 1, VerificationType: entity.EmailVerification, PlainToken: "token"}

	tests := []struct {
		name      string
		input     *event.UserVerificationCreated
		err       error
		setupMock func(m *mockTestUserVerificationCreatedUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email verification notify failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to notify"),
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.emailVerificationNotify.On("Exec", ctx, mock.MatchedBy(func(i *usecase.EmailVerificationNotifyUsecaseInput) bool {
					return i.Email == user.Email && i.Name == user.Name && i.Token == input.PlainToken
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to notify"))
			},
		},
		{
			name:  "reset password notify success",
			input: &event.UserVerificationCreated{UserID: user.ID, VerificationID: 1, VerificationType: entity.ResetPassword, PlainToken: "token"},
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.resetPasswordNotify.On("Exec", ctx, mock.MatchedBy(func(i *usecase.ResetPasswordNotifyUsecaseInput) bool {
					return i.Email == user.Email && i.Name == user.Name && i.Token == "token"
				})).Return(nil)
			},
		},
		{
			name:  "change email notify success",
			input: &event.UserVerificationCreated{UserID: user.ID, VerificationID: 1, VerificationType: entity.ChangeEmail, PlainToken: "token"},
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.changeEmailNotify.On("Exec", ctx, mock.MatchedBy(func(i *usecase.ChangeEmailNotifyUsecaseInput) bool {
					return i.Email == user.Email && i.Name == user.Name && i.Token == "token"
				})).Return(nil)
			},
		},
		{
			name:  "default notify success",
			input: &event.UserVerificationCreated{UserID: user.ID, VerificationID: 1, VerificationType: "", PlainToken: "token"},
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.emailVerificationNotify.On("Exec", ctx, mock.MatchedBy(func(i *usecase.EmailVerificationNotifyUsecaseInput) bool {
					return i.Email == user.Email && i.Name == user.Name && i.Token == "token"
				})).Return(nil)
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestUserVerificationCreatedUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.emailVerificationNotify.On("Exec", ctx, mock.MatchedBy(func(i *usecase.EmailVerificationNotifyUsecaseInput) bool {
					return i.Email == user.Email && i.Name == user.Name && i.Token == input.PlainToken
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserVerificationCreatedUsecase{
				userRepository:          mockrepository.NewMockUserRepository(t),
				emailVerificationNotify: mockusecase.NewMockEmailVerificationNotifyUsecase(t),
				resetPasswordNotify:     mockusecase.NewMockResetPasswordNotifyUsecase(t),
				changeEmailNotify:       mockusecase.NewMockChangeEmailNotifyUsecase(t),
				log:                     mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewUserVerificationCreatedUsecase(
				mock.userRepository,
				mock.emailVerificationNotify,
				mock.resetPasswordNotify,
				mock.changeEmailNotify,
				mock.log,
			)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

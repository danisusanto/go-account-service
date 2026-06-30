package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockhash "github.com/danisusanto/go-account-service/mocks/hash"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
)

type mockTestChangePasswordUsecase struct {
	userRepository *mockrepository.MockUserRepository
	hash           *mockhash.MockHash
	log            *mocklog.MockLog
}

func TestChangePasswordUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ChangePasswordUsecaseInput{
		UserID:          1,
		CurrentPassword: "Current123+",
		NewPassword:     "New123+",
	}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = input.UserID
	hashedNewPassword := "hashednewpassword"

	tests := []struct {
		name      string
		input     *usecase.ChangePasswordUsecaseInput
		err       error
		setupMock func(m *mockTestChangePasswordUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "current password is incorrect",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "the current password is incorrect"),
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("Compare", input.CurrentPassword, user.Password).Return(false)
			},
		},
		{
			name:  "failed to hash password",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash password"),
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("Compare", input.CurrentPassword, user.Password).Return(true)
				m.hash.On("Hash", input.NewPassword).Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash password"))
			},
		},
		{
			name:  "failed to set password",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set password"),
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("Compare", input.CurrentPassword, user.Password).Return(true)
				m.hash.On("Hash", input.NewPassword).Return(hashedNewPassword, nil)
				m.userRepository.On("SetPassword", ctx, user.ID, hashedNewPassword).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to set password"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestChangePasswordUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("Compare", input.CurrentPassword, user.Password).Return(true)
				m.hash.On("Hash", input.NewPassword).Return(hashedNewPassword, nil)
				m.userRepository.On("SetPassword", ctx, user.ID, hashedNewPassword).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangePasswordUsecase{
				userRepository: mockrepository.NewMockUserRepository(t),
				hash:           mockhash.NewMockHash(t),
				log:            mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewChangePasswordUsecase(mock.userRepository, mock.hash, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

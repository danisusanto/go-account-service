package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
)

type mockTestGetUserUsecase struct {
	userRepository *mockrepository.MockUserRepository
	log            *mocklog.MockLog
}

func TestGetUserUsecase(t *testing.T) {
	ctx := context.Background()

	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	tests := []struct {
		name      string
		input     int
		output    *usecase.GetUserUsecaseOutput
		err       error
		setupMock func(m *mockTestGetUserUsecase)
	}{
		{
			name:  "user not found",
			input: 1,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestGetUserUsecase) {
				m.userRepository.On("Get", ctx, 1).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: 2,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestGetUserUsecase) {
				m.userRepository.On("Get", ctx, 2).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "success",
			input: user.ID,
			output: &usecase.GetUserUsecaseOutput{
				ID:          user.ID,
				Name:        user.Name,
				Email:       user.Email,
				CountryCode: user.CountryCode,
				PhoneNumber: user.PhoneNumber,
				GoogleID:    user.GoogleID,
				IsActive:    user.IsUserActive(),
				Role:        user.Role,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			},
			setupMock: func(m *mockTestGetUserUsecase) {
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestGetUserUsecase{
				userRepository: mockrepository.NewMockUserRepository(t),
				log:            mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewGetUserUsecase(mock.userRepository, mock.log)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, res, tt.output)
		})
	}
}

package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUpdateUserUsecase struct {
	userRepository *mockrepository.MockUserRepository
	log            *mocklog.MockLog
}

func TestUpdateUserUsecase(t *testing.T) {
	ctx := context.Background()
	active := true
	input := &usecase.UpdateUserUsecaseInput{
		UserID:      1,
		Name:        "updated",
		CountryCode: "62",
		PhoneNumber: "62812345678",
		IsActive:    &active,
		Role:        entity.MANAGER,
		UpdatedBy:   2,
	}
	user := entity.NewUser(input.Name, "test@example.com", "hashedpassword", input.CountryCode, input.PhoneNumber)
	user.ID = input.UserID
	user.IsActive = input.IsActive
	user.Role = input.Role

	tests := []struct {
		name      string
		input     *usecase.UpdateUserUsecaseInput
		output    *usecase.UpdateUserUsecaseOutput
		err       error
		setupMock func(m *mockTestUpdateUserUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestUpdateUserUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "get user internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestUpdateUserUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "failed to update user",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to update user"),
			setupMock: func(m *mockTestUpdateUserUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.userRepository.On("Update", ctx, input.UserID, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to update user"))
			},
		},
		{
			name:  "success",
			input: input,
			output: &usecase.UpdateUserUsecaseOutput{
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
			setupMock: func(m *mockTestUpdateUserUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.userRepository.On("Update", ctx, input.UserID, mock.MatchedBy(func(u *entity.User) bool {
					return u.Name == input.Name &&
						u.CountryCode == input.CountryCode &&
						u.PhoneNumber == input.PhoneNumber &&
						u.IsActive == input.IsActive &&
						u.Role == input.Role &&
						u.UpdatedBy == input.UpdatedBy
				})).Return(user, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUpdateUserUsecase{
				userRepository: mockrepository.NewMockUserRepository(t),
				log:            mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewUpdateUserUsecase(mock.userRepository, mock.log)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, tt.output, res)
		})
	}
}

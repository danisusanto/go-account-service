package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	mockhash "github.com/dani-susanto/go-account-service/mocks/hash"
	mockjwt "github.com/dani-susanto/go-account-service/mocks/jwt"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockrepository "github.com/dani-susanto/go-account-service/mocks/repository"
	"github.com/dani-susanto/go-common/exception"
	commonjwt "github.com/dani-susanto/go-common/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestLoginUsecase struct {
	userRepository        *mockrepository.MockUserRepository
	userSessionRepository *mockrepository.MockUserSessionRepository
	hash                  *mockhash.MockHash
	accessTokenJWT        *mockjwt.MockJWT
	refreshTokenJWT       *mockjwt.MockJWT
}

func TestLoginUsecase(t *testing.T) {
	ctx := context.Background()

	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")

	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour))

	accessTokenOutput := &commonjwt.ClaimsOutput{
		Token:     "access_token",
		ID:        "access-id",
		ExpiresAt: expiresAt,
	}

	refreshTokenOutput := &commonjwt.ClaimsOutput{
		Token:     "refresh_token",
		ID:        "refresh-id",
		ExpiresAt: expiresAt,
	}

	tests := []struct {
		name      string
		input     *usecase.LoginUsecaseInput
		output    *usecase.LoginUsecaseOutput
		err       error
		setupMock func(m *mockTestLoginUsecase)
	}{
		{
			name: "user not found",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name: "internal server error",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name: "invalid credentials",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.UNAUTHORIZED, "invalid credentials"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(false)
			},
		},
		{
			name: "account is inactive",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.FORBIDDEN, "account is inactive"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(true)
				user.SetInactive()
			},
		},
		{
			name: "failed to generate access token jwt",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate access token jwt"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(true)
				user.SetActive()
				m.accessTokenJWT.On("Generate", mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate access token jwt"))
			},
		},
		{
			name: "failed to generate refresh token jwt",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(true)
				user.SetActive()
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"))
			},
		},
		{
			name: "failed to create session cache",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"),
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(true)
				user.SetActive()
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil).Once()
				m.refreshTokenJWT.On("Generate", mock.Anything).Return(refreshTokenOutput, nil).Once()
				m.userSessionRepository.On("CreateBulk", ctx, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"))
			},
		},
		{
			name: "success",
			input: &usecase.LoginUsecaseInput{
				Email:    "test@example.com",
				Password: "Test123+",
			},
			output: &usecase.LoginUsecaseOutput{
				AccessToken:  accessTokenOutput.Token,
				RefreshToken: refreshTokenOutput.Token,
			},
			setupMock: func(m *mockTestLoginUsecase) {
				m.userRepository.On("FindByEmail", ctx, mock.Anything).Return(user, nil)
				m.hash.On("Compare", mock.Anything, mock.Anything).Return(true)
				user.SetActive()
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil).Once()
				m.refreshTokenJWT.On("Generate", mock.Anything).Return(refreshTokenOutput, nil).Once()
				m.userSessionRepository.On("CreateBulk", ctx, mock.Anything).Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestLoginUsecase{
				userRepository:        mockrepository.NewMockUserRepository(t),
				userSessionRepository: mockrepository.NewMockUserSessionRepository(t),
				hash:                  mockhash.NewMockHash(t),
				accessTokenJWT:        mockjwt.NewMockJWT(t),
				refreshTokenJWT:       mockjwt.NewMockJWT(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewLoginUsecase(
				mock.userRepository,
				mock.userSessionRepository,
				mock.hash,
				mock.accessTokenJWT,
				mock.refreshTokenJWT,
				mocklog.NewMockLog(t),
			)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, res, tt.output)
		})
	}

}

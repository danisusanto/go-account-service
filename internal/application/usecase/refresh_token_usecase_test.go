package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockjwt "github.com/danisusanto/go-account-service/mocks/jwt"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	commonjwt "github.com/danisusanto/go-common/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestRefreshTokenUsecase struct {
	userRepository        *mockrepository.MockUserRepository
	userSessionRepository *mockrepository.MockUserSessionRepository
	accessTokenJWT        *mockjwt.MockJWT
	refreshTokenJWT       *mockjwt.MockJWT
}

func TestRefreshTokenUsecase(t *testing.T) {
	ctx := context.Background()
	token := "refresh_token"
	claims := &commonjwt.ClaimsInput{
		UserID: "1",
		RegisteredClaims: jwt.RegisteredClaims{
			ID: "refresh-id",
		},
	}
	session := entity.NewSession(1, claims.RegisteredClaims.ID, entity.RefreshToken, nil)
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = 1
	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour))
	accessTokenOutput := &commonjwt.ClaimsOutput{
		Token:     "new_access_token",
		ID:        "new-access-id",
		ExpiresAt: expiresAt,
	}
	refreshTokenOutput := &commonjwt.ClaimsOutput{
		Token:     "new_refresh_token",
		ID:        "new-refresh-id",
		ExpiresAt: expiresAt,
	}

	tests := []struct {
		name      string
		token     string
		output    *usecase.RefreshTokenOutput
		err       error
		setupMock func(m *mockTestRefreshTokenUsecase)
	}{
		{
			name:  "invalid refresh token",
			token: token,
			err:   exception.Throw(exception.BAD_REQUEST, "invalid token"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(nil, errors.New("invalid token"))
			},
		},
		{
			name:  "refresh token session expired",
			token: token,
			err:   exception.Throw(exception.BAD_REQUEST, "refresh token session expired"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "failed to find session",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to find session"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to find session"))
			},
		},
		{
			name:  "failed to delete session",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to delete session"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to delete session"))
			},
		},
		{
			name:  "invalid user id",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "strconv.Atoi: parsing \"invalid\": invalid syntax"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				invalidClaims := &commonjwt.ClaimsInput{
					UserID: "invalid",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "refresh-id",
					},
				}
				m.refreshTokenJWT.On("Validate", token).Return(invalidClaims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(invalidClaims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(invalidClaims.RegisteredClaims.ID)).Return(nil)
			},
		},
		{
			name:  "user not found",
			token: token,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "failed to get user",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to get user"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to get user"))
			},
		},
		{
			name:  "failed to generate access token jwt",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate accesst token jwt"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.accessTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.Name == user.Name &&
						i.Email == user.Email &&
						i.Role == string(user.Role)
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate access token jwt"))
			},
		},
		{
			name:  "failed to generate refresh token jwt",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.accessTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.Name == user.Name &&
						i.Email == user.Email &&
						i.Role == string(user.Role)
				})).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.RegisteredClaims.ID == accessTokenOutput.ID
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"))
			},
		},
		{
			name:  "failed to create session cache",
			token: token,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"),
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.Anything).Return(refreshTokenOutput, nil)
				m.userSessionRepository.On("CreateBulk", ctx, mock.MatchedBy(func(sessions []*entity.UserSession) bool {
					return len(sessions) == 2 &&
						sessions[0].UserID == user.ID &&
						sessions[0].SessionID == accessTokenOutput.ID &&
						sessions[0].SessionType == entity.AccessToken &&
						sessions[1].UserID == user.ID &&
						sessions[1].SessionID == refreshTokenOutput.ID &&
						sessions[1].SessionType == entity.RefreshToken
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"))
			},
		},
		{
			name:  "success",
			token: token,
			output: &usecase.RefreshTokenOutput{
				AccessToken:  accessTokenOutput.Token,
				RefreshToken: refreshTokenOutput.Token,
			},
			setupMock: func(m *mockTestRefreshTokenUsecase) {
				m.refreshTokenJWT.On("Validate", token).Return(claims, nil)
				m.userSessionRepository.On("FindByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(session, nil)
				m.userSessionRepository.On("DeleteByKey", ctx, entity.RefreshTokenKey(claims.RegisteredClaims.ID)).Return(nil)
				m.userRepository.On("Get", ctx, user.ID).Return(user, nil)
				m.accessTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.Name == user.Name &&
						i.Email == user.Email &&
						i.Role == string(user.Role)
				})).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.RegisteredClaims.ID == accessTokenOutput.ID
				})).Return(refreshTokenOutput, nil)
				m.userSessionRepository.On("CreateBulk", ctx, mock.MatchedBy(func(sessions []*entity.UserSession) bool {
					return len(sessions) == 2 &&
						sessions[0].UserID == user.ID &&
						sessions[0].SessionID == accessTokenOutput.ID &&
						sessions[0].SessionType == entity.AccessToken &&
						sessions[1].UserID == user.ID &&
						sessions[1].SessionID == refreshTokenOutput.ID &&
						sessions[1].SessionType == entity.RefreshToken
				})).Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestRefreshTokenUsecase{
				userRepository:        mockrepository.NewMockUserRepository(t),
				userSessionRepository: mockrepository.NewMockUserSessionRepository(t),
				accessTokenJWT:        mockjwt.NewMockJWT(t),
				refreshTokenJWT:       mockjwt.NewMockJWT(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewRefreshTokenUsecase(
				mock.userRepository,
				mock.userSessionRepository,
				mock.accessTokenJWT,
				mock.refreshTokenJWT,
				mocklog.NewMockLog(t),
			)

			res, err := usecase.Exec(ctx, tt.token)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, tt.output, res)
		})
	}
}

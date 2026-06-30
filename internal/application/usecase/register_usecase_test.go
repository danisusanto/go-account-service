package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	mockevent "github.com/danisusanto/go-account-service/mocks/event"
	mockhash "github.com/danisusanto/go-account-service/mocks/hash"
	mockjwt "github.com/danisusanto/go-account-service/mocks/jwt"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	commonjwt "github.com/danisusanto/go-common/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestRegisterUsecase struct {
	userRepository        *mockrepository.MockUserRepository
	userSessionRepository *mockrepository.MockUserSessionRepository
	event                 *mockevent.MockEvent
	accessTokenJWT        *mockjwt.MockJWT
	refreshTokenJWT       *mockjwt.MockJWT
	hash                  *mockhash.MockHash
}

func TestRegisterUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.RegisterUsecaseInput{
		Name:        "test",
		Email:       "test@example.com",
		Password:    "Test123+",
		CountryCode: "62",
		PhoneNumber: "62812345678",
	}
	existingUser := entity.NewUser(input.Name, input.Email, "hashedpassword", input.CountryCode, input.PhoneNumber)
	existingUser.ID = 1
	createdUser := entity.NewUser(input.Name, input.Email, "hashedpassword", input.CountryCode, input.PhoneNumber)
	createdUser.ID = 1
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
		input     *usecase.RegisterUsecaseInput
		output    *usecase.RegisterUsecaseOutput
		err       error
		setupMock func(m *mockTestRegisterUsecase)
	}{
		{
			name:  "failed to find user",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email already exists",
			input: input,
			err:   exception.Throw(exception.CONFLICT, "email already exists"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(existingUser, nil)
			},
		},
		{
			name:  "failed to generate password hashed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate password hashed"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash password"))
			},
		},
		{
			name:  "failed to create user",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create user"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.MatchedBy(func(u *entity.User) bool {
					return u.Name == input.Name &&
						u.Email == input.Email &&
						u.Password == "hashedpassword" &&
						u.CountryCode == input.CountryCode &&
						u.PhoneNumber == input.PhoneNumber
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create user"))
			},
		},
		{
			name:  "failed to publish user create event",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user create event"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.Anything).Return(createdUser, nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserCreated)
					return p.EventName == event.UserCreatedEvent &&
						ok &&
						data.UserID == createdUser.ID &&
						data.Email == createdUser.Email
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish event"))
			},
		},
		{
			name:  "failed to generate jwt token",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate jwt token"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.Anything).Return(createdUser, nil)
				m.event.On("Publish", ctx, mock.Anything).Return(nil)
				m.accessTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.Name == createdUser.Name &&
						i.Email == createdUser.Email &&
						i.Role == string(createdUser.Role)
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate access token jwt"))
			},
		},
		{
			name:  "failed to generate refresh token jwt",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.Anything).Return(createdUser, nil)
				m.event.On("Publish", ctx, mock.Anything).Return(nil)
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.RegisteredClaims.ID == accessTokenOutput.ID
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate refresh token jwt"))
			},
		},
		{
			name:  "failed to create session cache",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"),
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.Anything).Return(createdUser, nil)
				m.event.On("Publish", ctx, mock.Anything).Return(nil)
				m.accessTokenJWT.On("Generate", mock.Anything).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.Anything).Return(refreshTokenOutput, nil)
				m.userSessionRepository.On("CreateBulk", ctx, mock.MatchedBy(func(sessions []*entity.UserSession) bool {
					return len(sessions) == 2 &&
						sessions[0].UserID == createdUser.ID &&
						sessions[0].SessionID == accessTokenOutput.ID &&
						sessions[0].SessionType == entity.AccessToken &&
						sessions[1].UserID == createdUser.ID &&
						sessions[1].SessionID == refreshTokenOutput.ID &&
						sessions[1].SessionType == entity.RefreshToken
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create session cache"))
			},
		},
		{
			name:  "success",
			input: input,
			output: &usecase.RegisterUsecaseOutput{
				AccessToken:  accessTokenOutput.Token,
				RefreshToken: refreshTokenOutput.Token,
			},
			setupMock: func(m *mockTestRegisterUsecase) {
				m.userRepository.On("FindByEmail", ctx, input.Email).Return(nil, repository.ErrNotFound)
				m.hash.On("Hash", input.Password).Return("hashedpassword", nil)
				m.userRepository.On("Create", ctx, mock.MatchedBy(func(u *entity.User) bool {
					return u.Name == input.Name &&
						u.Email == input.Email &&
						u.Password == "hashedpassword" &&
						u.CountryCode == input.CountryCode &&
						u.PhoneNumber == input.PhoneNumber
				})).Return(createdUser, nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserCreated)
					return p.EventName == event.UserCreatedEvent &&
						ok &&
						data.UserID == createdUser.ID &&
						data.Email == createdUser.Email
				})).Return(nil)
				m.accessTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.Name == createdUser.Name &&
						i.Email == createdUser.Email &&
						i.Role == string(createdUser.Role)
				})).Return(accessTokenOutput, nil)
				m.refreshTokenJWT.On("Generate", mock.MatchedBy(func(i *commonjwt.ClaimsInput) bool {
					return i.UserID == "1" &&
						i.RegisteredClaims.ID == accessTokenOutput.ID
				})).Return(refreshTokenOutput, nil)
				m.userSessionRepository.On("CreateBulk", ctx, mock.MatchedBy(func(sessions []*entity.UserSession) bool {
					return len(sessions) == 2 &&
						sessions[0].UserID == createdUser.ID &&
						sessions[0].SessionID == accessTokenOutput.ID &&
						sessions[0].SessionType == entity.AccessToken &&
						sessions[1].UserID == createdUser.ID &&
						sessions[1].SessionID == refreshTokenOutput.ID &&
						sessions[1].SessionType == entity.RefreshToken
				})).Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestRegisterUsecase{
				userRepository:        mockrepository.NewMockUserRepository(t),
				userSessionRepository: mockrepository.NewMockUserSessionRepository(t),
				event:                 mockevent.NewMockEvent(t),
				accessTokenJWT:        mockjwt.NewMockJWT(t),
				refreshTokenJWT:       mockjwt.NewMockJWT(t),
				hash:                  mockhash.NewMockHash(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewRegisterUsecase(
				mock.userRepository,
				mock.userSessionRepository,
				mock.event,
				mock.accessTokenJWT,
				mock.refreshTokenJWT,
				mock.hash,
				mocklog.NewMockLog(t),
			)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}

			assert.Equal(t, tt.output, res)
		})
	}
}

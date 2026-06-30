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
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockrepository "github.com/danisusanto/go-account-service/mocks/repository"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestVerificationRequestUsecase struct {
	userRepository             *mockrepository.MockUserRepository
	userVerificationRepository *mockrepository.MockUserVerificationRepository
	event                      *mockevent.MockEvent
	hash                       *mockhash.MockHash
	log                        *mocklog.MockLog
}

func TestVerificationRequestUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.VerificationRequestUsecaseInput{
		UserID:           1,
		VerificationType: entity.EmailVerification,
		TokenType:        usecase.TextToken,
		TokenLength:      32,
		ExpiredDuration:  "2m",
		Payload:          "payload",
	}
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
	user.ID = input.UserID

	tests := []struct {
		name      string
		input     *usecase.VerificationRequestUsecaseInput
		output    *usecase.VerificationRequestUsecaseOutput
		err       error
		setupMock func(m *mockTestVerificationRequestUsecase)
	}{
		{
			name:  "user not found",
			input: input,
			err:   exception.Throw(exception.NOT_FOUND, "user not found"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, repository.ErrNotFound)
			},
		},
		{
			name:  "internal server error",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:  "email already verified",
			input: input,
			err:   exception.Throw(exception.BAD_REQUEST, "email already verified"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				verifiedUser := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "62812345678")
				verifiedUser.ID = input.UserID
				verifiedUser.SetEmailVerified()
				m.userRepository.On("Get", ctx, input.UserID).Return(verifiedUser, nil)
			},
		},
		{
			name: "failed to generate otp token",
			input: &usecase.VerificationRequestUsecaseInput{
				UserID:           1,
				VerificationType: entity.ResetPassword,
				TokenType:        usecase.OTPToken,
				TokenLength:      6,
				ExpiredDuration:  "2m",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate random token"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				i := &usecase.VerificationRequestUsecaseInput{UserID: 1, VerificationType: entity.ResetPassword, TokenType: usecase.OTPToken, TokenLength: 6, ExpiredDuration: "2m"}
				m.userRepository.On("Get", ctx, i.UserID).Return(user, nil)
				m.hash.On("GenerateOTP", i.TokenLength).Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate token"))
			},
		},
		{
			name:  "failed to generate text token",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate random token"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("GenerateToken", input.TokenLength).Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to generate token"))
			},
		},
		{
			name:  "failed to hash verification token",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash verification token"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("GenerateToken", input.TokenLength).Return("plain-token", nil)
				m.hash.On("Hash", "plain-token").Return("", exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to hash token"))
			},
		},
		{
			name: "failed parse verification expired duration",
			input: &usecase.VerificationRequestUsecaseInput{
				UserID:           1,
				VerificationType: entity.ResetPassword,
				TokenType:        usecase.TextToken,
				TokenLength:      32,
				ExpiredDuration:  "invalid",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed parse verification expired duration"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				i := &usecase.VerificationRequestUsecaseInput{UserID: 1, VerificationType: entity.ResetPassword, TokenType: usecase.TextToken, TokenLength: 32, ExpiredDuration: "invalid"}
				m.userRepository.On("Get", ctx, i.UserID).Return(user, nil)
				m.hash.On("GenerateToken", i.TokenLength).Return("plain-token", nil)
				m.hash.On("Hash", "plain-token").Return("tokenhash", nil)
			},
		},
		{
			name:  "failed to create user verification",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create user verification"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("GenerateToken", input.TokenLength).Return("plain-token", nil)
				m.hash.On("Hash", "plain-token").Return("tokenhash", nil)
				m.userVerificationRepository.On("Create", ctx, mock.MatchedBy(func(v *entity.UserVerification) bool {
					return v.UserID == user.ID &&
						v.VerificationType == input.VerificationType &&
						v.TokenHash == "tokenhash" &&
						v.Payload == input.Payload
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to create verification"))
			},
		},
		{
			name:  "failed to publish user verification created event",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish user verification created event"),
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				verification := entity.NewUserVerification(user.ID, input.VerificationType, "tokenhash", input.Payload, time.Now().Add(time.Minute))
				verification.ID = 1
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("GenerateToken", input.TokenLength).Return("plain-token", nil)
				m.hash.On("Hash", "plain-token").Return("tokenhash", nil)
				m.userVerificationRepository.On("Create", ctx, mock.Anything).Return(verification, nil)
				m.event.On("Publish", ctx, mock.MatchedBy(func(p *event.Payload) bool {
					data, ok := p.Data.(*event.UserVerificationCreated)
					return p.EventName == event.UserVerificationCreatedEvent &&
						ok &&
						data.UserID == user.ID &&
						data.VerificationID == verification.ID &&
						data.VerificationType == verification.VerificationType &&
						data.PlainToken == "plain-token"
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to publish event"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestVerificationRequestUsecase) {
				verification := entity.NewUserVerification(user.ID, input.VerificationType, "tokenhash", input.Payload, time.Now().Add(time.Minute))
				verification.ID = 1
				m.userRepository.On("Get", ctx, input.UserID).Return(user, nil)
				m.hash.On("GenerateToken", input.TokenLength).Return("plain-token", nil)
				m.hash.On("Hash", "plain-token").Return("tokenhash", nil)
				m.userVerificationRepository.On("Create", ctx, mock.Anything).Return(verification, nil)
				m.event.On("Publish", ctx, mock.Anything).Return(nil)
			},
			output: &usecase.VerificationRequestUsecaseOutput{
				PlainToken: "plain-token",
				User:       user,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestVerificationRequestUsecase{
				userRepository:             mockrepository.NewMockUserRepository(t),
				userVerificationRepository: mockrepository.NewMockUserVerificationRepository(t),
				event:                      mockevent.NewMockEvent(t),
				hash:                       mockhash.NewMockHash(t),
				log:                        mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewVerificationRequestUsecase(
				mock.userRepository,
				mock.userVerificationRepository,
				mock.event,
				mock.hash,
				mock.log,
			)

			res, err := usecase.Exec(ctx, tt.input)

			if err != nil {
				assert.Equal(t, tt.err, err)
			}
			if tt.output != nil {
				assert.Equal(t, tt.output.PlainToken, res.PlainToken)
				assert.Equal(t, tt.output.User, res.User)
				assert.NotNil(t, res.Verification)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

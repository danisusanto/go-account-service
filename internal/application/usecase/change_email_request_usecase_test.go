package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	mockenv "github.com/dani-susanto/go-account-service/mocks/env"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestChangeEmailRequestUsecase struct {
	verificationRequest *mockusecase.MockVerificationRequestUsecase
	env                 *mockenv.MockEnv
	log                 *mocklog.MockLog
}

func TestChangeEmailRequestUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ChangeEmailRequestUsecaseInput{
		UserID: 1,
		Email:  "test@example.com",
	}

	tests := []struct {
		name      string
		input     *usecase.ChangeEmailRequestUsecaseInput
		err       error
		setupMock func(m *mockTestChangeEmailRequestUsecase)
	}{
		{
			name:  "verification request failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"),
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.env.On("Get", "APP_CHANGE_EMAIL_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == input.UserID &&
						i.VerificationType == entity.ChangeEmail &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == input.Email
				})).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request verification"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestChangeEmailRequestUsecase) {
				m.env.On("Get", "APP_CHANGE_EMAIL_EXPIRED").Return("2m")
				m.verificationRequest.On("Exec", ctx, mock.MatchedBy(func(i *usecase.VerificationRequestUsecaseInput) bool {
					return i.UserID == input.UserID &&
						i.VerificationType == entity.ChangeEmail &&
						i.TokenType == usecase.OTPToken &&
						i.TokenLength == 6 &&
						i.ExpiredDuration == "2m" &&
						i.Payload == input.Email
				})).Return(&usecase.VerificationRequestUsecaseOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailRequestUsecase{
				verificationRequest: mockusecase.NewMockVerificationRequestUsecase(t),
				env:                 mockenv.NewMockEnv(t),
				log:                 mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewChangeEmailRequestUsecase(mock.verificationRequest, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

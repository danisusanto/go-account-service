package usecase_test

import (
	"context"
	"testing"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-account-service/internal/domain/notification/email"
	mockemail "github.com/dani-susanto/go-account-service/mocks/email"
	mockenv "github.com/dani-susanto/go-account-service/mocks/env"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	"github.com/dani-susanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestEmailVerificationNotifyUsecase struct {
	email *mockemail.MockEmail
	env   *mockenv.MockEnv
	log   *mocklog.MockLog
}

func TestEmailVerificationNotifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.EmailVerificationNotifyUsecaseInput{
		Email: "test@example.com",
		Name:  "test",
		Token: "token",
	}
	verificationURL := "https://example.com/verify-email"

	tests := []struct {
		name      string
		input     *usecase.EmailVerificationNotifyUsecaseInput
		err       error
		setupMock func(m *mockTestEmailVerificationNotifyUsecase)
	}{
		{
			name:  "send email failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email verification"),
			setupMock: func(m *mockTestEmailVerificationNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_EMAIL_VERIFICATION_URL").Return(verificationURL)
				m.env.On("Get", "APP_EMAIL_VERIFICATION_EXPIRED").Return("2m")
				m.email.On("Send", ctx, email.EmailVerificationSubject, input.Email, mock.MatchedBy(func(p *email.Payload) bool {
					data, ok := p.Data.(*email.EmailVerification)
					return p.TemplateName == email.EmailVerificationTemplate &&
						ok &&
						data.Name == input.Name &&
						data.VerificationURL == verificationURL+"?source="+input.Email+"&verify="+input.Token &&
						data.ExpiredIn == "2"
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestEmailVerificationNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_EMAIL_VERIFICATION_URL").Return(verificationURL)
				m.env.On("Get", "APP_EMAIL_VERIFICATION_EXPIRED").Return("2m")
				m.email.On("Send", ctx, email.EmailVerificationSubject, input.Email, mock.MatchedBy(func(p *email.Payload) bool {
					data, ok := p.Data.(*email.EmailVerification)
					return p.TemplateName == email.EmailVerificationTemplate &&
						ok &&
						data.Name == input.Name &&
						data.VerificationURL == verificationURL+"?source="+input.Email+"&verify="+input.Token &&
						data.ExpiredIn == "2"
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestEmailVerificationNotifyUsecase{
				email: mockemail.NewMockEmail(t),
				env:   mockenv.NewMockEnv(t),
				log:   mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewEmailVerificationNotifyUsecase(mock.email, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

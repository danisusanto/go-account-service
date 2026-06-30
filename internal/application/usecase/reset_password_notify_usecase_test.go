package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/notification/email"
	mockemail "github.com/danisusanto/go-account-service/mocks/email"
	mockenv "github.com/danisusanto/go-account-service/mocks/env"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestResetPasswordNotifyUsecase struct {
	email *mockemail.MockEmail
	env   *mockenv.MockEnv
	log   *mocklog.MockLog
}

func TestResetPasswordNotifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ResetPasswordNotifyUsecaseInput{
		Email: "test@example.com",
		Name:  "test",
		Token: "token",
	}

	tests := []struct {
		name      string
		input     *usecase.ResetPasswordNotifyUsecaseInput
		err       error
		setupMock func(m *mockTestResetPasswordNotifyUsecase)
	}{
		{
			name:  "send email failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send reset password email"),
			setupMock: func(m *mockTestResetPasswordNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_RESET_PASSWORD_EXPIRED").Return("2m")
				m.email.On("Send", ctx, email.ResetPasswordSubject, input.Email, mock.MatchedBy(func(p *email.Payload) bool {
					data, ok := p.Data.(*email.ResetPassword)
					return p.TemplateName == email.ResetPasswordTemplate &&
						ok &&
						data.Name == input.Name &&
						data.Token == input.Token &&
						data.ExpiredIn == "2"
				})).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestResetPasswordNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", "APP_RESET_PASSWORD_EXPIRED").Return("2m")
				m.email.On("Send", ctx, email.ResetPasswordSubject, input.Email, mock.MatchedBy(func(p *email.Payload) bool {
					data, ok := p.Data.(*email.ResetPassword)
					return p.TemplateName == email.ResetPasswordTemplate &&
						ok &&
						data.Name == input.Name &&
						data.Token == input.Token &&
						data.ExpiredIn == "2"
				})).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordNotifyUsecase{
				email: mockemail.NewMockEmail(t),
				env:   mockenv.NewMockEnv(t),
				log:   mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewResetPasswordNotifyUsecase(mock.email, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

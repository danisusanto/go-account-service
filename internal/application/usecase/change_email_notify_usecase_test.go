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

type mockTestChangeEmailNotifyUsecase struct {
	email *mockemail.MockEmail
	env   *mockenv.MockEnv
	log   *mocklog.MockLog
}

func TestChangeEmailNotifyUsecase(t *testing.T) {
	ctx := context.Background()
	input := &usecase.ChangeEmailNotifyUsecaseInput{
		Email: "test@example.com",
		Name:  "test",
		Token: "token",
	}

	tests := []struct {
		name      string
		input     *usecase.ChangeEmailNotifyUsecaseInput
		err       error
		setupMock func(m *mockTestChangeEmailNotifyUsecase)
	}{
		{
			name: "send email failed",
			input: &usecase.ChangeEmailNotifyUsecaseInput{
				Email: "test@example.com",
				Name:  "test",
				Token: "token",
			},
			err: exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send change email token"),
			setupMock: func(m *mockTestChangeEmailNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", mock.Anything).Return("2m")
				m.email.On("Send", ctx, email.ChangeEmailSubject, input.Email, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to send email"))
			},
		},
		{
			name: "success",
			input: &usecase.ChangeEmailNotifyUsecaseInput{
				Email: "test@example.com",
				Name:  "test",
				Token: "token",
			},
			setupMock: func(m *mockTestChangeEmailNotifyUsecase) {
				m.log.On("Info", mock.Anything, mock.Anything)
				m.env.On("Get", mock.Anything).Return("2m")
				m.email.On("Send", ctx, email.ChangeEmailSubject, input.Email, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailNotifyUsecase{
				email: mockemail.NewMockEmail(t),
				env:   mockenv.NewMockEnv(t),
				log:   mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewChangeEmailNotifyUsecase(mock.email, mock.env, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

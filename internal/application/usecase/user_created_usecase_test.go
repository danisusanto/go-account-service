package usecase_test

import (
	"context"
	"testing"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-account-service/internal/domain/event"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/stretchr/testify/assert"
)

type mockTestUserCreatedUsecase struct {
	emailVerificationRequest *mockusecase.MockEmailVerificationRequestUsecase
	log                      *mocklog.MockLog
}

func TestUserCreatedUsecase(t *testing.T) {
	ctx := context.Background()
	input := &event.UserCreated{UserID: 1, Email: "test@example.com"}

	tests := []struct {
		name      string
		input     *event.UserCreated
		err       error
		setupMock func(m *mockTestUserCreatedUsecase)
	}{
		{
			name:  "email verification request failed",
			input: input,
			err:   exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request email verification"),
			setupMock: func(m *mockTestUserCreatedUsecase) {
				m.emailVerificationRequest.On("Exec", ctx, input.UserID).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "failed to request email verification"))
			},
		},
		{
			name:  "success",
			input: input,
			setupMock: func(m *mockTestUserCreatedUsecase) {
				m.emailVerificationRequest.On("Exec", ctx, input.UserID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserCreatedUsecase{
				emailVerificationRequest: mockusecase.NewMockEmailVerificationRequestUsecase(t),
				log:                      mocklog.NewMockLog(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			usecase := usecase.NewUserCreatedUsecase(mock.emailVerificationRequest, mock.log)

			err := usecase.Exec(ctx, tt.input)

			assert.Equal(t, tt.err, err)
		})
	}
}

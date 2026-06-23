package event

import (
	"testing"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/domain/event"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUserResetPasswordHandler struct {
	userResetPasswordUsecase *mockusecase.MockUserResetPasswordUsecase
}

func TestUserResetPasswordHandler(t *testing.T) {
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	tests := []struct {
		name      string
		payload   event.Payload
		wantErr   bool
		setupMock func(m *mockTestUserResetPasswordHandler)
	}{
		{
			name: "success",
			payload: event.Payload{
				EventName: event.UserResetPasswordEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: false,
			setupMock: func(m *mockTestUserResetPasswordHandler) {
				m.userResetPasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "usecase error",
			payload: event.Payload{
				EventName: event.UserResetPasswordEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: true,
			setupMock: func(m *mockTestUserResetPasswordHandler) {
				m.userResetPasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserResetPasswordHandler{
				userResetPasswordUsecase: mockusecase.NewMockUserResetPasswordUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			userResetPasswordHandler := NewUserResetPasswordHandler(
				handler,
				mock.userResetPasswordUsecase,
				jsonUtil,
				log,
			)
			err := userResetPasswordHandler.UserResetPasswordHandler(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserResetPasswordHandlerTransform(t *testing.T) {
	userResetPasswordUsecase := mockusecase.NewMockUserResetPasswordUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userResetPasswordHandler := NewUserResetPasswordHandler(
		handler,
		userResetPasswordUsecase,
		jsonUtil,
		log,
	)

	tests := []struct {
		name       string
		payload    event.Payload
		wantUserID int
	}{
		{
			name: "transform success",
			payload: event.Payload{
				EventName: event.UserResetPasswordEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantUserID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userResetPasswordHandler.Transform(tt.payload)

			assert.Equal(t, tt.wantUserID, result.UserID)
		})
	}
}

func TestUserResetPasswordHandlerRegister(t *testing.T) {
	userResetPasswordUsecase := mockusecase.NewMockUserResetPasswordUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userResetPasswordHandler := NewUserResetPasswordHandler(
		handler,
		userResetPasswordUsecase,
		jsonUtil,
		log,
	)

	eventName := event.UserResetPasswordEvent
	userResetPasswordHandler.Register(eventName)

	assert.NotNil(t, handler.handlers[eventName])
	assert.Equal(t, 1, len(handler.handlers))
}

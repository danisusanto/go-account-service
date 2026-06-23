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

type mockTestUserEmailVerifiedHandler struct {
	userEmailVerifiedUsecase *mockusecase.MockUserEmailVerifiedUsecase
}

func TestUserEmailVerifiedHandler(t *testing.T) {
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	tests := []struct {
		name      string
		payload   event.Payload
		wantErr   bool
		setupMock func(m *mockTestUserEmailVerifiedHandler)
	}{
		{
			name: "success",
			payload: event.Payload{
				EventName: event.UserEmailVerifiedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: false,
			setupMock: func(m *mockTestUserEmailVerifiedHandler) {
				m.userEmailVerifiedUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "usecase error",
			payload: event.Payload{
				EventName: event.UserEmailVerifiedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: true,
			setupMock: func(m *mockTestUserEmailVerifiedHandler) {
				m.userEmailVerifiedUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserEmailVerifiedHandler{
				userEmailVerifiedUsecase: mockusecase.NewMockUserEmailVerifiedUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			userEmailVerifiedHandler := NewUserEmailVerifiedHandler(
				handler,
				mock.userEmailVerifiedUsecase,
				jsonUtil,
				log,
			)

			err := userEmailVerifiedHandler.UserEmailVerifiedHandler(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserEmailVerifiedHandlerTransform(t *testing.T) {
	userEmailVerifiedUsecase := mockusecase.NewMockUserEmailVerifiedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userEmailVerifiedHandler := NewUserEmailVerifiedHandler(
		handler,
		userEmailVerifiedUsecase,
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
				EventName: event.UserEmailVerifiedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantUserID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userEmailVerifiedHandler.Transform(tt.payload)

			assert.Equal(t, tt.wantUserID, result.UserID)
		})
	}
}

func TestUserEmailVerifiedHandlerRegister(t *testing.T) {
	userEmailVerifiedUsecase := mockusecase.NewMockUserEmailVerifiedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userEmailVerifiedHandler := NewUserEmailVerifiedHandler(
		handler,
		userEmailVerifiedUsecase,
		jsonUtil,
		log,
	)

	eventName := event.UserEmailVerifiedEvent
	userEmailVerifiedHandler.Register(eventName)

	assert.NotNil(t, handler.handlers[eventName])
	assert.Equal(t, 1, len(handler.handlers))
}

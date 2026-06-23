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

type mockTestUserCreatedHandler struct {
	userCreatedUsecase *mockusecase.MockUserCreatedUsecase
}

func TestUserCreatedHandler(t *testing.T) {
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	tests := []struct {
		name      string
		payload   event.Payload
		wantErr   bool
		setupMock func(m *mockTestUserCreatedHandler)
	}{
		{
			name: "success",
			payload: event.Payload{
				EventName: event.UserCreatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
					"email":   "test@example.com",
				},
			},
			wantErr: false,
			setupMock: func(m *mockTestUserCreatedHandler) {
				m.userCreatedUsecase.
					On("Exec", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "usecase error",
			payload: event.Payload{
				EventName: event.UserCreatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
					"email":   "test@example.com",
				},
			},
			wantErr: true,
			setupMock: func(m *mockTestUserCreatedHandler) {
				m.userCreatedUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserCreatedHandler{
				userCreatedUsecase: mockusecase.NewMockUserCreatedUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			userCreatedHandler := NewUserCreatedHandler(
				handler,
				mock.userCreatedUsecase,
				jsonUtil,
				log,
			)

			err := userCreatedHandler.UserCreatedHandler(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserCreatedHandlerTransform(t *testing.T) {
	userCreatedUsecase := mockusecase.NewMockUserCreatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userCreatedHandler := NewUserCreatedHandler(
		handler,
		userCreatedUsecase,
		jsonUtil,
		log,
	)

	tests := []struct {
		name       string
		payload    event.Payload
		wantUserID int
		wantEmail  string
	}{
		{
			name: "transform success",
			payload: event.Payload{
				EventName: event.UserCreatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
					"email":   "test@example.com",
				},
			},
			wantUserID: 1,
			wantEmail:  "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCreatedHandler.Transform(tt.payload)

			assert.Equal(t, tt.wantUserID, result.UserID)
			assert.Equal(t, tt.wantEmail, result.Email)
		})
	}
}

func TestUserCreatedHandlerRegister(t *testing.T) {
	userCreatedUsecase := mockusecase.NewMockUserCreatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userCreatedHandler := NewUserCreatedHandler(
		handler,
		userCreatedUsecase,
		jsonUtil,
		log,
	)

	eventName := event.UserCreatedEvent
	userCreatedHandler.Register(eventName)

	assert.NotNil(t, handler.handlers[eventName])
	assert.Equal(t, 1, len(handler.handlers))
}

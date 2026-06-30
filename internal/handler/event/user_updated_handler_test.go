package event

import (
	"testing"

	"github.com/danisusanto/go-account-service/internal/domain/event"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/json"
	"github.com/danisusanto/go-common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUserUpdatedHandler struct {
	userUpdatedUsecase *mockusecase.MockUserUpdatedUsecase
}

func TestUserUpdatedHandler(t *testing.T) {
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	tests := []struct {
		name      string
		payload   event.Payload
		wantErr   bool
		setupMock func(m *mockTestUserUpdatedHandler)
	}{
		{
			name: "success",
			payload: event.Payload{
				EventName: event.UserUpdatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: false,
			setupMock: func(m *mockTestUserUpdatedHandler) {
				m.userUpdatedUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "usecase error",
			payload: event.Payload{
				EventName: event.UserUpdatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantErr: true,
			setupMock: func(m *mockTestUserUpdatedHandler) {
				m.userUpdatedUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserUpdatedHandler{
				userUpdatedUsecase: mockusecase.NewMockUserUpdatedUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			userUpdatedHandler := NewUserUpdatedHandler(
				handler,
				mock.userUpdatedUsecase,
				jsonUtil,
				log,
			)

			err := userUpdatedHandler.UserUpdatedHandler(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUpdatedHandlerTransform(t *testing.T) {
	userUpdatedUsecase := mockusecase.NewMockUserUpdatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userUpdatedHandler := NewUserUpdatedHandler(
		handler,
		userUpdatedUsecase,
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
				EventName: event.UserUpdatedEvent,
				Data: map[string]interface{}{
					"user_id": 1,
				},
			},
			wantUserID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userUpdatedHandler.Transform(tt.payload)

			assert.Equal(t, tt.wantUserID, result.UserID)
		})
	}
}

func TestUserUpdatedHandlerRegister(t *testing.T) {
	userUpdatedUsecase := mockusecase.NewMockUserUpdatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userUpdatedHandler := NewUserUpdatedHandler(
		handler,
		userUpdatedUsecase,
		jsonUtil,
		log,
	)

	eventName := event.UserUpdatedEvent
	userUpdatedHandler.Register(eventName)

	assert.NotNil(t, handler.handlers[eventName])
	assert.Equal(t, 1, len(handler.handlers))
}

package event

import (
	"testing"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/event"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUserVerificationCreatedHandler struct {
	userVerificationCreatedUsecase *mockusecase.MockUserVerificationCreatedUsecase
}

func TestUserVerificationCreatedHandler(t *testing.T) {
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	tests := []struct {
		name      string
		payload   event.Payload
		wantErr   bool
		setupMock func(m *mockTestUserVerificationCreatedHandler)
	}{
		{
			name: "success",
			payload: event.Payload{
				EventName: event.UserVerificationCreatedEvent,
				Data: map[string]interface{}{
					"UserID":           1,
					"VerificationID":   1,
					"VerificationType": entity.EmailVerification,
					"PlainToken":       "token123",
				},
			},
			wantErr: false,
			setupMock: func(m *mockTestUserVerificationCreatedHandler) {
				m.userVerificationCreatedUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "usecase error",
			payload: event.Payload{
				EventName: event.UserVerificationCreatedEvent,
				Data: map[string]interface{}{
					"UserID":           1,
					"VerificationID":   1,
					"VerificationType": entity.EmailVerification,
					"PlainToken":       "token123",
				},
			},
			wantErr: true,
			setupMock: func(m *mockTestUserVerificationCreatedHandler) {
				m.userVerificationCreatedUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUserVerificationCreatedHandler{
				userVerificationCreatedUsecase: mockusecase.NewMockUserVerificationCreatedUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			userVerificationCreatedHandler := NewUserVerificationCreatedHandler(
				handler,
				mock.userVerificationCreatedUsecase,
				jsonUtil,
				log,
			)

			err := userVerificationCreatedHandler.UserVerificationCreatedHandler(tt.payload)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserVerificationCreatedHandlerTransform(t *testing.T) {
	userVerificationCreatedUsecase := mockusecase.NewMockUserVerificationCreatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userVerificationCreatedHandler := NewUserVerificationCreatedHandler(
		handler,
		userVerificationCreatedUsecase,
		jsonUtil,
		log,
	)

	tests := []struct {
		name                 string
		payload              event.Payload
		wantUserID           int
		wantVerificationID   int
		wantVerificationType entity.VerificationType
	}{
		{
			name: "transform success",
			payload: event.Payload{
				EventName: event.UserVerificationCreatedEvent,
				Data: map[string]interface{}{
					"UserID":           1,
					"VerificationID":   1,
					"VerificationType": entity.EmailVerification,
					"PlainToken":       "token123",
				},
			},
			wantUserID:           1,
			wantVerificationID:   1,
			wantVerificationType: entity.EmailVerification,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userVerificationCreatedHandler.Transform(tt.payload)

			assert.Equal(t, tt.wantUserID, result.UserID)
			assert.Equal(t, tt.wantVerificationID, result.VerificationID)
			assert.Equal(t, tt.wantVerificationType, result.VerificationType)
		})
	}
}

func TestUserVerificationCreatedHandlerRegister(t *testing.T) {
	userVerificationCreatedUsecase := mockusecase.NewMockUserVerificationCreatedUsecase(t)
	jsonUtil := json.New()
	log := log.New()

	handler := &Handler{
		handlers: make(map[string]func(event.Payload) error),
	}

	userVerificationCreatedHandler := NewUserVerificationCreatedHandler(
		handler,
		userVerificationCreatedUsecase,
		jsonUtil,
		log,
	)

	eventName := event.UserVerificationCreatedEvent
	userVerificationCreatedHandler.Register(eventName)

	assert.NotNil(t, handler.handlers[eventName])
	assert.Equal(t, 1, len(handler.handlers))
}

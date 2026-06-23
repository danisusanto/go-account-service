package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dani-susanto/go-common/jwt"
	"github.com/dani-susanto/go-common/log"

	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestEmailVerificationRequestHandler struct {
	emailVerificationRequestUsecase *mockusecase.MockEmailVerificationRequestUsecase
}

func TestEmailVerificationRequestHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestEmailVerificationRequestHandler)
	}{
		{
			name:       "unauthorized",
			claims:     nil,
			wantStatus: http.StatusUnauthorized,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "unauthorized",
			},
		},
		{
			name:       "usecase error",
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestEmailVerificationRequestHandler) {
				m.emailVerificationRequestUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestEmailVerificationRequestHandler) {
				m.emailVerificationRequestUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestEmailVerificationRequestHandler{
				emailVerificationRequestUsecase: mockusecase.NewMockEmailVerificationRequestUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewEmailVerificationRequestHandler(
				handler{
					json:      jsonUtil,
					responder: commonrest.NewResponder(jsonUtil),
					log:       log.New(),
				},
				mock.emailVerificationRequestUsecase,
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/email-verification", nil)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.EmailVerificationRequestHandler(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrorResponse != nil {
				var got commonrest.ErrorResponse
				jsonUtil.Decode(w.Body.Bytes(), &got)

				assert.Equal(t, tt.wantErrorResponse.Status, got.Status)
				assert.Equal(t, tt.wantErrorResponse.Message, got.Message)

				if tt.wantErrorResponse.Errors != nil {
					assert.Equal(t, tt.wantErrorResponse.Errors, got.Errors)
				}
			}
		})
	}
}

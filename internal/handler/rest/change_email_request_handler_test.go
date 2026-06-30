package rest

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danisusanto/go-common/jwt"
	"github.com/danisusanto/go-common/log"

	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/json"
	commonrest "github.com/danisusanto/go-common/rest"
	"github.com/danisusanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestChangeEmailRequestHandler struct {
	changeEmailRequestUsecase *mockusecase.MockChangeEmailRequestUsecase
}

func TestChangeEmailRequestHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestChangeEmailRequestHandler)
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
			name:       "invalid request body",
			body:       "",
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "invalid request body",
			},
		},
		{
			name:       "validation - email missing",
			body:       &changeEmailRequest{},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email": {
						"The email field is required.",
						"The email field must be a valid email address.",
					},
				},
			},
		},
		{
			name:       "validation - invalid email format",
			body:       &changeEmailRequest{Email: "invalid"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email": {"The email field must be a valid email address."},
				},
			},
		},
		{
			name:       "usecase error",
			body:       &changeEmailRequest{Email: "test@example.com"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestChangeEmailRequestHandler) {
				m.changeEmailRequestUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &changeEmailRequest{Email: "test@example.com"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestChangeEmailRequestHandler) {
				m.changeEmailRequestUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailRequestHandler{
				changeEmailRequestUsecase: mockusecase.NewMockChangeEmailRequestUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewChangeEmailRequestHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.changeEmailRequestUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/email/request", &body)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.ChangeEmailRequestHandler(w, r)

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

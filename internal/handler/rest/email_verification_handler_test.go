package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dani-susanto/go-common/log"

	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/dani-susanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestEmailVerificationHandler struct {
	emailVerificationUsecase *mockusecase.MockEmailVerificationUsecase
}

func TestEmailVerificationHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		email               string
		token               string
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestEmailVerificationHandler)
	}{
		{
			name:       "validation - email missing",
			email:      "",
			token:      "123456",
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
			name:       "validation - token missing",
			email:      "test@example.com",
			token:      "",
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"token": {"The token field is required."},
				},
			},
		},
		{
			name:       "validation - invalid email format",
			email:      "invalid",
			token:      "123456",
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
			name:       "validation - all fields missing",
			email:      "",
			token:      "",
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email": {
						"The email field is required.",
						"The email field must be a valid email address.",
					},
					"token": {"The token field is required."},
				},
			},
		},
		{
			name:       "usecase error",
			email:      "test@example.com",
			token:      "123456",
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestEmailVerificationHandler) {
				m.emailVerificationUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			email:      "test@example.com",
			token:      "123456",
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "your email has been verified successfully.",
			},
			setupMock: func(m *mockTestEmailVerificationHandler) {
				m.emailVerificationUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestEmailVerificationHandler{
				emailVerificationUsecase: mockusecase.NewMockEmailVerificationUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewEmailVerificationHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.emailVerificationUsecase,
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/verify?source="+tt.email+"&verify="+tt.token, nil)
			handler.EmailVerificationHandler(w, r)

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

			if tt.wantSuccessResponse != nil {
				var got commonrest.SuccessResponse
				jsonUtil.Decode(w.Body.Bytes(), &got)

				assert.Equal(t, tt.wantSuccessResponse.Status, got.Status)
				assert.Equal(t, tt.wantSuccessResponse.Message, got.Message)
			}
		})
	}
}

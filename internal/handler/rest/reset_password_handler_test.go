package rest

import (
	"bytes"
	gojson "encoding/json"
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

type mockTestResetPasswordHandler struct {
	resetPasswordUsecase *mockusecase.MockResetPasswordUsecase
}

func TestResetPasswordHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestResetPasswordHandler)
	}{
		{
			name:       "invalid request body",
			body:       "",
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "invalid request body",
			},
		},
		{
			name:       "validation - all fields missing",
			body:       &resetPasswordRequest{},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email": {
						"The email field is required.",
						"The email field must be a valid email address.",
					},
					"token": {
						"The token field is required.",
						"The token field must be a 6-digit OTP.",
					},
					"new_password": {
						"The new_password field is required.",
						"Password must be at least 8 characters.",
						"Password must contain at least one uppercase letter.",
						"Password must contain at least one lowercase letter.",
						"Password must contain at least one number.",
						"Password must contain at least one special character.",
					},
					"new_password_confirm": {"The new_password_confirm field is required."},
				},
			},
		},
		{
			name:       "validation - invalid token format",
			body:       &resetPasswordRequest{Email: "test@example.com", Token: "invalid", NewPassword: "NewPass123+", NewPasswordConfirm: "NewPass123+"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"token": {"The token field must be a 6-digit OTP."},
				},
			},
		},
		{
			name:       "validation - invalid new password format",
			body:       &resetPasswordRequest{Email: "test@example.com", Token: "123456", NewPassword: "weak", NewPasswordConfirm: "weak"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"new_password": {
						"Password must be at least 8 characters.",
						"Password must contain at least one uppercase letter.",
						"Password must contain at least one number.",
						"Password must contain at least one special character.",
					},
				},
			},
		},
		{
			name:       "validation - password confirm mismatch",
			body:       &resetPasswordRequest{Email: "test@example.com", Token: "123456", NewPassword: "NewPass123+", NewPasswordConfirm: "Different123+"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"new_password_confirm": {"The new_password_confirm field confirmation does not match."},
				},
			},
		},
		{
			name:       "usecase error",
			body:       &resetPasswordRequest{Email: "test@example.com", Token: "123456", NewPassword: "NewPass123+", NewPasswordConfirm: "NewPass123+"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestResetPasswordHandler) {
				m.resetPasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &resetPasswordRequest{Email: "test@example.com", Token: "123456", NewPassword: "NewPass123+", NewPasswordConfirm: "NewPass123+"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestResetPasswordHandler) {
				m.resetPasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordHandler{
				resetPasswordUsecase: mockusecase.NewMockResetPasswordUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewResetPasswordHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.resetPasswordUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/reset-password", &body)
			handler.ResetPasswordHandler(w, r)

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

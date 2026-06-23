package rest

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dani-susanto/go-common/jwt"
	"github.com/dani-susanto/go-common/log"

	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/dani-susanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestChangePasswordHandler struct {
	changePasswordUsecase *mockusecase.MockChangePasswordUsecase
}

func TestChangePasswordHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestChangePasswordHandler)
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
			name:       "validation - all fields missing",
			body:       &changePasswordRequest{},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"current_password": {"The current_password field is required."},
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
			name:       "validation - invalid new password format",
			body:       &changePasswordRequest{CurrentPassword: "Test123+", NewPassword: "weak", NewPasswordConfirm: "weak"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &changePasswordRequest{CurrentPassword: "Test123+", NewPassword: "NewPass123+", NewPasswordConfirm: "Different123+"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &changePasswordRequest{CurrentPassword: "Test123+", NewPassword: "NewPass123+", NewPasswordConfirm: "NewPass123+"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestChangePasswordHandler) {
				m.changePasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &changePasswordRequest{CurrentPassword: "Test123+", NewPassword: "NewPass123+", NewPasswordConfirm: "NewPass123+"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestChangePasswordHandler) {
				m.changePasswordUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangePasswordHandler{
				changePasswordUsecase: mockusecase.NewMockChangePasswordUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewChangePasswordHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.changePasswordUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/change-password", &body)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.ChangePasswordHandler(w, r)

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

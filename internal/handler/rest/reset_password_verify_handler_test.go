package rest

import (
	"bytes"
	gojson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danisusanto/go-common/log"

	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/json"
	commonrest "github.com/danisusanto/go-common/rest"
	"github.com/danisusanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestResetPasswordVerifyHandler struct {
	resetPasswordVerifyUsecase *mockusecase.MockResetPasswordVerifyUsecase
}

func TestResetPasswordVerifyHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestResetPasswordVerifyHandler)
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
			body:       &resetPasswordVerifyRequest{},
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
				},
			},
		},
		{
			name:       "validation - token missing",
			body:       &resetPasswordVerifyRequest{Email: "test@example.com"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"token": {
						"The token field is required.",
						"The token field must be a 6-digit OTP.",
					},
				},
			},
		},
		{
			name:       "validation - invalid token format",
			body:       &resetPasswordVerifyRequest{Email: "test@example.com", Token: "invalid"},
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
			name:       "usecase error",
			body:       &resetPasswordVerifyRequest{Email: "test@example.com", Token: "123456"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestResetPasswordVerifyHandler) {
				m.resetPasswordVerifyUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &resetPasswordVerifyRequest{Email: "test@example.com", Token: "123456"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestResetPasswordVerifyHandler) {
				m.resetPasswordVerifyUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestResetPasswordVerifyHandler{
				resetPasswordVerifyUsecase: mockusecase.NewMockResetPasswordVerifyUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewResetPasswordVerifyHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.resetPasswordVerifyUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/reset-password/verify", &body)
			handler.ResetPasswordVerifyHandler(w, r)

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

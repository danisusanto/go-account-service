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

type mockTestChangeEmailVerifyHandler struct {
	changeEmailVerifyUsecase *mockusecase.MockChangeEmailVerifyUsecase
}

func TestChangeEmailVerifyHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestChangeEmailVerifyHandler)
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
			name:       "validation - token missing",
			body:       &changeEmailVerifyRequest{},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &changeEmailVerifyRequest{Token: "1"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &changeEmailVerifyRequest{Token: "123456"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestChangeEmailVerifyHandler) {
				m.changeEmailVerifyUsecase.On("Exec", mock.Anything, mock.Anything).Return(exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &changeEmailVerifyRequest{Token: "123456"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusNoContent,
			setupMock: func(m *mockTestChangeEmailVerifyHandler) {
				m.changeEmailVerifyUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestChangeEmailVerifyHandler{
				changeEmailVerifyUsecase: mockusecase.NewMockChangeEmailVerifyUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewChangeEmailVerifyHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.changeEmailVerifyUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/email/verify", &body)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.ChangeEmailVerifyHandler(w, r)

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

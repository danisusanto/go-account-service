package rest

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	mockusecase "github.com/dani-susanto/go-account-service/mocks/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/json"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/dani-susanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestLoginHandler struct {
	loginUsecase *mockusecase.MockLoginUsecase
}

func TestLoginHandler(t *testing.T) {
	ctx := context.Background()
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestLoginHandler)
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
			body:       &loginRequest{},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email": {
						"The email field is required.",
						"The email field must be a valid email address.",
					},
					"password": {"The password field is required."},
				},
			},
		},
		{
			name:       "validation - invalid email format",
			body:       &loginRequest{Email: "test"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"email":    {"The email field must be a valid email address."},
					"password": {"The password field is required."},
				},
			},
		},
		{
			name:       "validation - password missing",
			body:       &loginRequest{Email: "test@example.com"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"password": {"The password field is required."},
				},
			},
		},
		{
			name:       "usecase error",
			body:       &loginRequest{Email: "test@example.com", Password: "Test123+"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestLoginHandler) {
				m.loginUsecase.On("Exec", ctx, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &loginRequest{Email: "test@example.com", Password: "Test123+"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "login successful",
				Data: &usecase.LoginUsecaseOutput{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				},
			},
			setupMock: func(m *mockTestLoginHandler) {
				m.loginUsecase.On("Exec", ctx, mock.Anything).Return(&usecase.LoginUsecaseOutput{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestLoginHandler{
				loginUsecase: mockusecase.NewMockLoginUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewLoginHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.loginUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/login", &body)
			handler.LoginHandler(w, r)

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

				if got.Status {
					raw, err := jsonUtil.Encode(got.Data)
					if err != nil {
						t.Fatal(err)
					}
					var output *usecase.LoginUsecaseOutput
					jsonUtil.Decode([]byte(raw), &output)
					assert.Equal(t, tt.wantSuccessResponse.Data, output)
				}
			}
		})
	}
}

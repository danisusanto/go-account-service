package rest

import (
	"bytes"
	gojson "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danisusanto/go-common/log"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/json"
	commonrest "github.com/danisusanto/go-common/rest"
	"github.com/danisusanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestRegisterHandler struct {
	registerUsecase *mockusecase.MockRegisterUsecase
}

func TestRegisterHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestRegisterHandler)
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
			body:       &registerRequest{},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"name": {"The name field is required."},
					"email": {
						"The email field is required.",
						"The email field must be a valid email address.",
					},
					"password": {
						"The password field is required.",
						"Password must be at least 8 characters.",
						"Password must contain at least one uppercase letter.",
						"Password must contain at least one lowercase letter.",
						"Password must contain at least one number.",
						"Password must contain at least one special character.",
					},
					"country_code": {
						"The country_code field is required.",
						"The country_code field must be a valid country code (digits only, 1-3 length).",
					},
					"phone_number": {
						"The phone_number field is required.",
						"The phone_number field must be a valid phone number.",
					},
				},
			},
		},
		{
			name:       "validation - invalid email format",
			body:       &registerRequest{Name: "John Doe", Email: "invalid", Password: "Pass123+", CountryCode: "62", PhoneNumber: "62812345678"},
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
			name:       "validation - invalid password format",
			body:       &registerRequest{Name: "John Doe", Email: "test@example.com", Password: "weak", CountryCode: "62", PhoneNumber: "62812345678"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"password": {
						"Password must be at least 8 characters.",
						"Password must contain at least one uppercase letter.",
						"Password must contain at least one number.",
						"Password must contain at least one special character.",
					},
				},
			},
		},
		{
			name:       "validation - invalid country code",
			body:       &registerRequest{Name: "John Doe", Email: "test@example.com", Password: "Pass123+", CountryCode: "invalid", PhoneNumber: "62812345678"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"country_code": {"The country_code field must be a valid country code (digits only, 1-3 length)."},
				},
			},
		},
		{
			name:       "validation - invalid phone number",
			body:       &registerRequest{Name: "John Doe", Email: "test@example.com", Password: "Pass123+", CountryCode: "62", PhoneNumber: "123"},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"phone_number": {"The phone_number field must be a valid phone number."},
				},
			},
		},
		{
			name:       "usecase error",
			body:       &registerRequest{Name: "John Doe", Email: "test@example.com", Password: "Pass123+", CountryCode: "62", PhoneNumber: "62812345678"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestRegisterHandler) {
				m.registerUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &registerRequest{Name: "John Doe", Email: "test@example.com", Password: "Pass123+", CountryCode: "62", PhoneNumber: "62812345678"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "register successful",
				Data: &usecase.RegisterUsecaseOutput{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				},
			},
			setupMock: func(m *mockTestRegisterHandler) {
				m.registerUsecase.On("Exec", mock.Anything, mock.Anything).Return(&usecase.RegisterUsecaseOutput{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestRegisterHandler{
				registerUsecase: mockusecase.NewMockRegisterUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewRegisterHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.registerUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/register", &body)
			handler.RegisterHandler(w, r)

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
					var output *usecase.RegisterUsecaseOutput
					jsonUtil.Decode([]byte(raw), &output)
					assert.Equal(t, tt.wantSuccessResponse.Data, output)
				}
			}
		})
	}
}

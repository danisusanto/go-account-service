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

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	mockusecase "github.com/danisusanto/go-account-service/mocks/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/json"
	commonrest "github.com/danisusanto/go-common/rest"
	"github.com/danisusanto/go-common/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestUpdateMeHandler struct {
	updateUserUsecase *mockusecase.MockUpdateUserUsecase
}

func TestUpdateMeHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestUpdateMeHandler)
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
			name:       "validation - invalid country code",
			body:       &updateMeRequest{Name: "John", CountryCode: "invalid", PhoneNumber: ""},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &updateMeRequest{Name: "", CountryCode: "62", PhoneNumber: "123"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
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
			body:       &updateMeRequest{Name: "John Doe", CountryCode: "62", PhoneNumber: "62812345678"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestUpdateMeHandler) {
				m.updateUserUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success - all fields",
			body:       &updateMeRequest{Name: "John Doe", CountryCode: "62", PhoneNumber: "62812345678"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "update user successful",
				Data: &usecase.UpdateUserUsecaseOutput{
					ID:   1,
					Name: "John Doe",
				},
			},
			setupMock: func(m *mockTestUpdateMeHandler) {
				m.updateUserUsecase.On("Exec", mock.Anything, mock.Anything).Return(&usecase.UpdateUserUsecaseOutput{ID: 1, Name: "John Doe"}, nil)
			},
		},
		{
			name:       "success - partial fields",
			body:       &updateMeRequest{Name: "John Doe"},
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "update user successful",
				Data: &usecase.UpdateUserUsecaseOutput{
					ID:   1,
					Name: "John Doe",
				},
			},
			setupMock: func(m *mockTestUpdateMeHandler) {
				m.updateUserUsecase.On("Exec", mock.Anything, mock.Anything).Return(&usecase.UpdateUserUsecaseOutput{ID: 1, Name: "John Doe"}, nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestUpdateMeHandler{
				updateUserUsecase: mockusecase.NewMockUpdateUserUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewUpdateMeHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.updateUserUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, "/me", &body)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.UpdateMeHandler(w, r)

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
					var output *usecase.UpdateUserUsecaseOutput
					jsonUtil.Decode([]byte(raw), &output)
					assert.Equal(t, tt.wantSuccessResponse.Data, output)
				}
			}
		})
	}
}

package rest

import (
	"bytes"
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

type mockTestRefreshTokenHandler struct {
	refreshTokenUsecase *mockusecase.MockRefreshTokenUsecase
}

func TestRefreshTokenHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestRefreshTokenHandler)
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
			name:       "validation - refresh_token missing",
			body:       &refreshTokenRequest{},
			wantStatus: http.StatusBadRequest,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "request validation failed",
				Errors: map[string][]string{
					"refresh_token": {"The refresh_token field is required."},
				},
			},
		},
		{
			name:       "usecase error",
			body:       &refreshTokenRequest{RefreshToken: "invalid_token"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestRefreshTokenHandler) {
				m.refreshTokenUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			body:       &refreshTokenRequest{RefreshToken: "valid_refresh_token"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "refreshToken successful",
				Data: &usecase.RefreshTokenOutput{
					AccessToken: "new_access_token",
				},
			},
			setupMock: func(m *mockTestRefreshTokenHandler) {
				m.refreshTokenUsecase.On("Exec", mock.Anything, mock.Anything).Return(&usecase.RefreshTokenOutput{AccessToken: "new_access_token"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestRefreshTokenHandler{
				refreshTokenUsecase: mockusecase.NewMockRefreshTokenUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewRefreshTokenHandler(
				handler{
					validation: validator.New(),
					json:       jsonUtil,
					responder:  commonrest.NewResponder(jsonUtil),
					log:        log.New(),
				},
				mock.refreshTokenUsecase,
			)

			var body bytes.Buffer
			if err := gojson.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/refresh-token", &body)
			handler.RefreshTokenHandler(w, r)

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
					var output *usecase.RefreshTokenOutput
					jsonUtil.Decode([]byte(raw), &output)
					assert.Equal(t, tt.wantSuccessResponse.Data, output)
				}
			}
		})
	}
}

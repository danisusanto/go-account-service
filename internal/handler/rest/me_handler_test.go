package rest

import (
	"context"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTestMeHandler struct {
	getUserUsecase *mockusecase.MockGetUserUsecase
}

func TestMeHandler(t *testing.T) {
	jsonUtil := json.New()

	tests := []struct {
		name                string
		body                any
		claims              *jwt.ClaimsInput
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
		wantErrorResponse   *commonrest.ErrorResponse
		setupMock           func(m *mockTestMeHandler)
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
			name:       "usecase error",
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusInternalServerError,
			wantErrorResponse: &commonrest.ErrorResponse{
				Status:  false,
				Message: "internal server error",
			},
			setupMock: func(m *mockTestMeHandler) {
				m.getUserUsecase.On("Exec", mock.Anything, mock.Anything).Return(nil, exception.Throw(exception.INTERNAL_SERVER_ERROR, "internal server error"))
			},
		},
		{
			name:       "success",
			claims:     &jwt.ClaimsInput{UserID: "1"},
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "get user successful",
				Data: &usecase.GetUserUsecaseOutput{
					ID: 1,
				},
			},
			setupMock: func(m *mockTestMeHandler) {
				m.getUserUsecase.On("Exec", mock.Anything, mock.Anything).Return(&usecase.GetUserUsecaseOutput{ID: 1}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockTestMeHandler{
				getUserUsecase: mockusecase.NewMockGetUserUsecase(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(&mock)
			}

			handler := NewMeHandler(
				handler{
					json:      jsonUtil,
					responder: commonrest.NewResponder(jsonUtil),
					log:       log.New(),
				},
				mock.getUserUsecase,
			)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/me", nil)
			if tt.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, tt.claims))
			}
			handler.MeHandler(w, r)

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
					var output *usecase.GetUserUsecaseOutput
					jsonUtil.Decode([]byte(raw), &output)
					assert.Equal(t, tt.wantSuccessResponse.Data, output)
				}
			}
		})
	}
}

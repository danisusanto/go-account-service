package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dani-susanto/go-common/log"

	"github.com/dani-susanto/go-common/json"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	jsonUtil := json.New()

	handler := NewHealthHandler(
		handler{
			json:      jsonUtil,
			responder: commonrest.NewResponder(jsonUtil),
			log:       log.New(),
		},
	)

	tests := []struct {
		name                string
		wantStatus          int
		wantSuccessResponse *commonrest.SuccessResponse
	}{
		{
			name:       "success",
			wantStatus: http.StatusOK,
			wantSuccessResponse: &commonrest.SuccessResponse{
				Status:  true,
				Message: "all systems operational",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/health", nil)
			handler.HealthHandler(w, r)

			var got commonrest.SuccessResponse
			jsonUtil.Decode(w.Body.Bytes(), &got)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantSuccessResponse.Status, got.Status)
			assert.Equal(t, tt.wantSuccessResponse.Message, got.Message)
		})
	}
}

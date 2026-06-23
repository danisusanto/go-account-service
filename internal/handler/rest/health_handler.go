package rest

import (
	"net/http"
)

type healthHandler struct {
	handler
}

func NewHealthHandler(handler handler) *healthHandler {
	return &healthHandler{
		handler: handler,
	}
}

// HealthHandler
//
//	@Summary		Service health check
//	@Description	Check the availability and operational status of the service.
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	rest.SuccessResponse	"Service is healthy"
//	@Router			/health [get]
func (h *healthHandler) Register() *healthHandler {
	h.router.Get("/health", h.HealthHandler)
	return h
}

func (h *healthHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	h.responder.Success(w, http.StatusOK, "OK", "all systems operational", nil)
}

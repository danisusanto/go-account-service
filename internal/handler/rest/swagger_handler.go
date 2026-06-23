package rest

import (
	"github.com/dani-susanto/go-account-service/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type swaggerHandler struct {
	handler
}

func NewSwaggerHandler(handler handler) *swaggerHandler {
	return &swaggerHandler{
		handler: handler,
	}
}

func (h *swaggerHandler) Register() *swaggerHandler {
	docs.SwaggerInfo.Host = h.env.Get("SWAGGER_BASE_URL")
	h.router.Get("/swagger/{path...}", httpSwagger.WrapHandler)
	return h
}

// func (h *swaggerHandler) SwaggerHandler(w http.ResponseWriter, r *http.Request) {
// w.Header().Set("Content-Type", "application/json")
// w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
// }

package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type loginHandler struct {
	handler
	login usecase.LoginUsecase
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewLoginHandler(
	handler handler,
	login usecase.LoginUsecase,
) *loginHandler {
	return &loginHandler{
		handler: handler,
		login:   login,
	}
}

// LoginHandler
//
//	@Summary		User authentication
//	@Description	Authenticate a user using email and password credentials. Returns an access token and refresh token for subsequent authenticated requests.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginRequest	true	"User login credentials"
//	@Success		200		{object}	rest.SuccessResponse{data=usecase.LoginUsecaseOutput}
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload or validation failure"
//	@Failure		401		{object}	rest.ErrorResponse	"Invalid email or password"
//	@Failure		403		{object}	rest.ErrorResponse	"Account is inactive or suspended"
//	@Failure		404		{object}	rest.ErrorResponse	"User account not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/login [post]
func (h *loginHandler) Register() *loginHandler {
	path := h.buildPath("/login")
	h.router.Post(
		path,
		h.LoginHandler,
		h.router.ExcludeGlobalRateLimiter("POST", path),
		h.router.RateLimiter(5, "1m"),
	)
	return h
}

func (h *loginHandler) Validate(req loginRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()
	validation.Set("password", req.Password).Required()

	return validation.Validate()
}

func (h *loginHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	res, err := h.login.Exec(ctx, (*usecase.LoginUsecaseInput)(&req))
	if err != nil {
		h.log.Error(
			"login failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusOK, res, "login successful", nil)
}

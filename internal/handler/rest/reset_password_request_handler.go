package rest

import (
	"net/http"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-common/exception"
)

type resetPasswordRequestHandler struct {
	handler
	resetPasswordRequest usecase.ResetPasswordRequestUsecase
}

type resetPasswordDemandRequest struct {
	Email string `json:"email"`
}

func NewResetPasswordRequestHandler(
	handler handler,
	resetPasswordRequest usecase.ResetPasswordRequestUsecase,
) *resetPasswordRequestHandler {
	return &resetPasswordRequestHandler{
		handler:              handler,
		resetPasswordRequest: resetPasswordRequest,
	}
}

// ResetPasswordRequestHandler
//
//	@Summary		Request password reset
//	@Description	Initiate a password reset process by sending a password reset verification code to the user's registered email address.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body	resetPasswordDemandRequest	true	"Password reset request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload or validation failure"
//	@Failure		404		{object}	rest.ErrorResponse	"User not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/reset-password/request [post]
func (h *resetPasswordRequestHandler) Register() *resetPasswordRequestHandler {
	path := h.buildPath("/reset-password/request")
	h.router.Post(
		path,
		h.ResetPasswordRequestHandler,
		h.router.ExcludeGlobalRateLimiter("POST", path),
		h.router.RateLimiter(1, "2m"))
	return h
}

func (h *resetPasswordRequestHandler) Validate(req resetPasswordDemandRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()

	return validation.Validate()
}

func (h *resetPasswordRequestHandler) ResetPasswordRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordDemandRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	err := h.resetPasswordRequest.Exec(ctx, req.Email)
	if err != nil {
		h.log.Error(
			"reset password request failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

package rest

import (
	"net/http"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-common/exception"
)

type resetPasswordVerifyHandler struct {
	handler
	resetPasswordVerify usecase.ResetPasswordVerifyUsecase
}

type resetPasswordVerifyRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func NewResetPasswordVerifyHandler(handler handler, resetPasswordVerify usecase.ResetPasswordVerifyUsecase) *resetPasswordVerifyHandler {
	return &resetPasswordVerifyHandler{
		handler:             handler,
		resetPasswordVerify: resetPasswordVerify,
	}
}

// ResetPasswordVerifyHandler
//
//	@Summary		Verify password reset request
//	@Description	Verify a password reset token before proceeding with the password reset process.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body	resetPasswordVerifyRequest	true	"Password reset verification request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload, validation failure, or invalid verification token"
//	@Failure		404		{object}	rest.ErrorResponse	"User not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/reset-password/verify [post]
func (h *resetPasswordVerifyHandler) Register() *resetPasswordVerifyHandler {
	h.router.Post(h.buildPath("/reset-password/verify"), h.ResetPasswordVerifyHandler)
	return h
}

func (h *resetPasswordVerifyHandler) Validate(req resetPasswordVerifyRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()
	validation.Set("token", req.Token).Required().OTP()

	return validation.Validate()
}

func (h *resetPasswordVerifyHandler) ResetPasswordVerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordVerifyRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	err := h.resetPasswordVerify.Exec(
		ctx,
		&usecase.ResetPasswordVerifyUsecaseInput{
			Email: req.Email,
			Token: req.Token,
		},
	)
	if err != nil {
		h.log.Error(
			"reset password verify failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

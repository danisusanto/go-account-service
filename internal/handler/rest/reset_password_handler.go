package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type resetPasswordHandler struct {
	handler
	resetPassword usecase.ResetPasswordUsecase
}

type resetPasswordRequest struct {
	Email              string `json:"email"`
	Token              string `json:"token"`
	NewPassword        string `json:"new_password"`
	NewPasswordConfirm string `json:"new_password_confirm"`
}

func NewResetPasswordHandler(
	handler handler,
	resetPassword usecase.ResetPasswordUsecase,
) *resetPasswordHandler {
	return &resetPasswordHandler{
		handler:       handler,
		resetPassword: resetPassword,
	}
}

// ResetPasswordHandler
//
//	@Summary		Reset password
//	@Description	Reset the authenticated user's password using a valid password reset verification token.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body	resetPasswordRequest	true	"Password reset request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload, validation failure, or invalid password reset token"
//	@Failure		401		{object}	rest.ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	rest.ErrorResponse	"User not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/reset-password [post]
func (h *resetPasswordHandler) Register() *resetPasswordHandler {
	h.router.Post(h.buildPath("/reset-password"), h.ResetPasswordHandler)
	return h
}

func (h *resetPasswordHandler) Validate(req resetPasswordRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()
	validation.Set("token", req.Token).Required().OTP()
	validation.Set("new_password", req.NewPassword).Required().Password()
	validation.Set("new_password_confirm", req.NewPasswordConfirm).Required().PasswordConfirm(req.NewPassword)

	return validation.Validate()
}

func (h *resetPasswordHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	err := h.resetPassword.Exec(ctx, &usecase.ResetPasswordUsecaseInput{
		Email:       req.Email,
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		h.log.Error(
			"reset password failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

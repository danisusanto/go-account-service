package rest

import (
	"net/http"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-common/exception"
)

type changePasswordHandler struct {
	handler
	changePassword usecase.ChangePasswordUsecase
}

type changePasswordRequest struct {
	CurrentPassword    string `json:"current_password"`
	NewPassword        string `json:"new_password"`
	NewPasswordConfirm string `json:"new_password_confirm"`
}

func NewChangePasswordHandler(
	handler handler,
	changePassword usecase.ChangePasswordUsecase,
) *changePasswordHandler {
	return &changePasswordHandler{
		handler:        handler,
		changePassword: changePassword,
	}
}

// ChangePasswordHandler
//
//	@Summary		Change password
//	@Description	Update the authenticated user's password by providing the current password and a new password.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body	changePasswordRequest	true	"Password change request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload, validation failure, or incorrect current password"
//	@Failure		401		{object}	rest.ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	rest.ErrorResponse	"User not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/change-password [post]
func (h *changePasswordHandler) Register() *changePasswordHandler {
	h.router.Post(h.buildPath("/change-password"), h.ChangePasswordHandler, h.router.Authenticated)
	return h
}

func (h *changePasswordHandler) Validate(req changePasswordRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("current_password", req.CurrentPassword).Required()
	validation.Set("new_password", req.NewPassword).Required().Password()
	validation.Set("new_password_confirm", req.NewPasswordConfirm).Required().PasswordConfirm(req.NewPassword)

	return validation.Validate()
}

func (h *changePasswordHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	var req changePasswordRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	err := h.changePassword.Exec(ctx, &usecase.ChangePasswordUsecaseInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		h.log.Error(
			"change password failed failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

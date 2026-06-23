package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type changeEmailVerifyHandler struct {
	handler
	changeEmailVerify usecase.ChangeEmailVerifyUsecase
}

type changeEmailVerifyRequest struct {
	Token string `json:"token"`
}

func NewChangeEmailVerifyHandler(
	handler handler,
	changeEmailVerify usecase.ChangeEmailVerifyUsecase,
) *changeEmailVerifyHandler {
	return &changeEmailVerifyHandler{
		handler:           handler,
		changeEmailVerify: changeEmailVerify,
	}
}

// ChangeEmailVerifyHandler
//
//	@Summary		Verify email change
//	@Description	Verify the email change request using the provided verification code and update the authenticated user's email address.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body	changeEmailVerifyRequest	true	"Email change verification"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload or validation failure"
//	@Failure		401		{object}	rest.ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	rest.ErrorResponse	"User not found"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/change-email/verify [post]
func (h *changeEmailVerifyHandler) Register() *changeEmailVerifyHandler {
	h.router.Post(h.buildPath("/change-email/verify"), h.ChangeEmailVerifyHandler, h.router.Authenticated)
	return h
}

func (h *changeEmailVerifyHandler) Validate(req changeEmailVerifyRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("token", req.Token).Required().OTP()

	return validation.Validate()
}

func (h *changeEmailVerifyHandler) ChangeEmailVerifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	var req changeEmailVerifyRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	if err := h.changeEmailVerify.Exec(
		ctx,
		&usecase.ChangeEmailVerifyUsecaseInput{
			UserID: userID,
			Token:  req.Token,
		},
	); err != nil {
		h.log.Error(
			"change email failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

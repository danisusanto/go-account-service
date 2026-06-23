package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type changeEmailRequestHandler struct {
	handler
	changeEmailRequest usecase.ChangeEmailRequestUsecase
}

type changeEmailRequest struct {
	Email string `json:"email"`
}

func NewChangeEmailRequestHandler(
	handler handler,
	changeEmailRequest usecase.ChangeEmailRequestUsecase,
) *changeEmailRequestHandler {
	return &changeEmailRequestHandler{
		handler:            handler,
		changeEmailRequest: changeEmailRequest,
	}
}

// ChangeEmailRequestHandler
//
//	@Summary		Request email change verification
//	@Description	Initiate an email change process by sending a verification code to the specified email address. The verification code must be used to complete the email change operation.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body	changeEmailRequest	true	"Email change request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	rest.ErrorResponse	"Invalid request payload or validation failure"
//	@Failure		401		{object}	rest.ErrorResponse	"Unauthorized"
//	@Failure		500		{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/change-email/request [post]
func (h *changeEmailRequestHandler) Register() *changeEmailRequestHandler {
	h.router.Post(h.buildPath("/change-email/request"), h.ChangeEmailRequestHandler, h.router.Authenticated)
	return h
}

func (h *changeEmailRequestHandler) Validate(req changeEmailRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()

	return validation.Validate()
}

func (h *changeEmailRequestHandler) ChangeEmailRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	var req changeEmailRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	if err := h.changeEmailRequest.Exec(
		ctx,
		&usecase.ChangeEmailRequestUsecaseInput{
			UserID: userID,
			Email:  req.Email,
		},
	); err != nil {
		h.log.Error(
			"change email request failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

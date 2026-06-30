package rest

import (
	"net/http"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-common/exception"
)

type emailVerificationRequestHandler struct {
	handler
	emailVerificationRequest usecase.EmailVerificationRequestUsecase
}

func NewEmailVerificationRequestHandler(
	handler handler,
	emailVerificationRequest usecase.EmailVerificationRequestUsecase,
) *emailVerificationRequestHandler {
	return &emailVerificationRequestHandler{
		handler:                  handler,
		emailVerificationRequest: emailVerificationRequest,
	}
}

// EmailVerificationRequestHandler
//
//	@Summary		Request email verification
//	@Description	Send a new email verification link to the authenticated user's email address.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		204	"No Content"
//	@Failure		401	{object}	rest.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	rest.ErrorResponse	"Unexpected internal server error"
//	@Router			/v1/email-verification/request [post]
func (h *emailVerificationRequestHandler) Register() *emailVerificationRequestHandler {
	path := h.buildPath("/email-verification/request")
	h.router.Post(
		path,
		h.EmailVerificationRequestHandler,
		h.router.Authenticated,
		h.router.ExcludeGlobalRateLimiter("POST", path),
		h.router.RateLimiter(1, "2m"),
	)
	return h
}

func (h *emailVerificationRequestHandler) EmailVerificationRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.emailVerificationRequest.Exec(ctx, userID)
	if err != nil {
		h.log.Error(
			"email verification request failed",
			"error", err,
			"user_id", userID,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusNoContent, nil, "", nil)
}

package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type emailVerificationHandler struct {
	handler
	emailVerification usecase.EmailVerificationUsecase
}

type emailVerificationRequest struct {
	Email string `json:"email"`
	Token string `json:"password"`
}

func NewEmailVerificationHandler(handler handler, emailVerification usecase.EmailVerificationUsecase) *emailVerificationHandler {
	return &emailVerificationHandler{
		handler:           handler,
		emailVerification: emailVerification,
	}
}

// EmailVerificationHandler
//
//	@Summary		Verify email address
//	@Description	Verify a user's email address using the verification token provided in the email verification link.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			source	query		string					true	"Email address"
//	@Param			verify	query		string					true	"Email verification token"
//	@Success		200		{object}	rest.SuccessResponse	"Email verified successfully"
//	@Failure		400		{object}	rest.ErrorResponse		"Invalid request parameters or validation failure"
//	@Failure		404		{object}	rest.ErrorResponse		"User not found"
//	@Failure		500		{object}	rest.ErrorResponse		"Unexpected internal server error"
//	@Router			/v1/email-verification/verify [get]
func (h *emailVerificationHandler) Register() *emailVerificationHandler {
	h.router.Get(h.buildPath("/email-verification/verify"), h.EmailVerificationHandler)
	return h
}

func (h *emailVerificationHandler) Validate(req emailVerificationRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("email", req.Email).Required().Email()
	validation.Set("token", req.Token).Required()

	return validation.Validate()
}

func (h *emailVerificationHandler) EmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	req := emailVerificationRequest{
		Email: r.URL.Query().Get("source"),
		Token: r.URL.Query().Get("verify"),
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	err := h.emailVerification.Exec(ctx, (*usecase.EmailVerificationUsecaseInput)(&req))
	if err != nil {
		h.log.Error(
			"email verification failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusOK, "OK", "your email has been verified successfully.", nil)
}

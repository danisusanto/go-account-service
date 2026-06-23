package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type meHandler struct {
	handler
	user usecase.GetUserUsecase
}

func NewMeHandler(
	handler handler,
	user usecase.GetUserUsecase,
) *meHandler {
	return &meHandler{
		handler: handler,
		user:    user,
	}
}

// MeHandler
//
//	@Summary		Get authenticated user
//	@Description	Retrieve profile information for the currently authenticated user.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	rest.SuccessResponse{data=usecase.GetUserUsecaseOutput}	"User profile retrieved successfully"
//	@Failure		401	{object}	rest.ErrorResponse										"Unauthorized"
//	@Failure		404	{object}	rest.ErrorResponse										"User not found"
//	@Failure		500	{object}	rest.ErrorResponse										"Unexpected internal server error"
//	@Router			/v1/me [get]
func (h *meHandler) Register() *meHandler {
	h.router.Get(h.buildPath("/me"), h.MeHandler, h.router.Authenticated)
	return h
}

func (h *meHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	res, err := h.user.Exec(ctx, userID)
	if err != nil {
		h.log.Error(
			"get user failed",
			"error", err,
			"user_id", userID,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusOK, res, "get user successful", nil)
}

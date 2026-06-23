package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
)

type refreshTokenHandler struct {
	handler
	refreshToken usecase.RefreshTokenUsecase
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshTokenHandler(
	handler handler,
	refreshToken usecase.RefreshTokenUsecase,
) *refreshTokenHandler {
	return &refreshTokenHandler{
		handler:      handler,
		refreshToken: refreshToken,
	}
}

// RefreshTokenHandler
//
//	@Summary		Refresh access token
//	@Description	Generate a new access token and refresh token using a valid refresh token. The previous refresh token will be invalidated after successful use.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		refreshTokenRequest										true	"Refresh token request"
//	@Success		200		{object}	rest.SuccessResponse{data=usecase.RefreshTokenOutput}	"Token refreshed successfully"
//	@Failure		400		{object}	rest.ErrorResponse										"Invalid request payload, invalid refresh token, expired refresh token, or refresh token session expired"
//	@Failure		404		{object}	rest.ErrorResponse										"User not found"
//	@Failure		500		{object}	rest.ErrorResponse										"Unexpected internal server error"
//	@Router			/v1/refresh-token [post]
func (h *refreshTokenHandler) Register() {
	h.router.Post(h.buildPath("/refresh-token"), h.RefreshTokenHandler)
}

func (h *refreshTokenHandler) Validate(req refreshTokenRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("refresh_token", req.RefreshToken).Required()

	return validation.Validate()
}

func (h *refreshTokenHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	res, err := h.refreshToken.Exec(ctx, req.RefreshToken)
	if err != nil {
		h.log.Error(
			"refresh token failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusOK, res, "refreshToken successful", nil)
}

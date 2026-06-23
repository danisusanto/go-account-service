package rest

import (
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/application/usecase"
	"github.com/dani-susanto/go-common/exception"
	"github.com/dani-susanto/go-common/format"
)

type updateMeHandler struct {
	handler
	updateUser usecase.UpdateUserUsecase
}

type updateMeRequest struct {
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	PhoneNumber string `json:"phone_number"`
}

func NewUpdateMeHandler(
	handler handler,
	updateUser usecase.UpdateUserUsecase,
) *updateMeHandler {
	return &updateMeHandler{
		handler:    handler,
		updateUser: updateUser,
	}
}

// UpdateMeHandler
//
//	@Summary		Update profile
//	@Description	Update the authenticated user's profile information, including name and phone number.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		updateMeRequest												true	"Profile update request"
//	@Success		200		{object}	rest.SuccessResponse{data=usecase.UpdateUserUsecaseOutput}	"Profile updated successfully"
//	@Failure		400		{object}	rest.ErrorResponse											"Invalid request payload or validation failure"
//	@Failure		401		{object}	rest.ErrorResponse											"Unauthorized"
//	@Failure		404		{object}	rest.ErrorResponse											"User not found"
//	@Failure		500		{object}	rest.ErrorResponse											"Unexpected internal server error"
//	@Router			/v1/me [patch]
func (h *updateMeHandler) Register() *updateMeHandler {
	h.router.Patch(h.buildPath("/me"), h.UpdateMeHandler, h.router.Authenticated)
	return h
}

func (h *updateMeHandler) Validate(req updateMeRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("name", req.Name).Nullable()
	validation.Set("country_code", req.CountryCode).Nullable().CountryCode()
	validation.Set("phone_number", req.PhoneNumber).Nullable().Phone()

	return validation.Validate()
}

func (h *updateMeHandler) UpdateMeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractUserID(w, r)
	if !ok {
		return
	}

	var req updateMeRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	res, err := h.updateUser.Exec(
		ctx,
		&usecase.UpdateUserUsecaseInput{
			UserID:      userID,
			Name:        req.Name,
			CountryCode: req.CountryCode,
			PhoneNumber: format.NormalizePhone(req.CountryCode, req.PhoneNumber),
			UpdatedBy:   userID,
		},
	)
	if err != nil {
		h.log.Error(
			"update user failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}
	h.responder.Success(w, http.StatusOK, res, "update user successful", nil)
}

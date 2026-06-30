package rest

import (
	"net/http"

	"github.com/danisusanto/go-account-service/internal/application/usecase"
	"github.com/danisusanto/go-common/exception"
	"github.com/danisusanto/go-common/format"
)

type registerHandler struct {
	handler
	register usecase.RegisterUsecase
}

type registerRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CountryCode string `json:"country_code"`
	PhoneNumber string `json:"phone_number"`
}

func NewRegisterHandler(
	handler handler,
	register usecase.RegisterUsecase,
) *registerHandler {
	return &registerHandler{
		handler:  handler,
		register: register,
	}
}

// RegisterHandler
//
//	@Summary		User registration
//	@Description	Create a new user account and issue access and refresh tokens for authenticated API access.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerRequest												true	"User registration information"
//	@Success		200		{object}	rest.SuccessResponse{data=usecase.RegisterUsecaseOutput}	"Registration successful"
//	@Failure		400		{object}	rest.ErrorResponse											"Invalid request payload or validation failure"
//	@Failure		409		{object}	rest.ErrorResponse											"Email address already exists"
//	@Failure		500		{object}	rest.ErrorResponse											"Unexpected internal server error"
//	@Router			/v1/register [post]
func (h *registerHandler) Register() *registerHandler {
	h.router.Post(h.buildPath("/register"), h.RegisterHandler)
	return h
}

func (h *registerHandler) Validate(req registerRequest) map[string][]string {
	validation := h.validation.Prepare()

	validation.Set("name", req.Name).Required()
	validation.Set("email", req.Email).Required().Email()
	validation.Set("password", req.Password).Required().Password()
	validation.Set("country_code", req.CountryCode).Required().CountryCode()
	validation.Set("phone_number", req.PhoneNumber).Required().Phone()

	return validation.Validate()
}

func (h *registerHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := h.json.DecodeHttp(r, &req); err != nil {
		h.responder.Error(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if errs := h.Validate(req); len(errs) > 0 {
		h.responder.Error(w, http.StatusBadRequest, "request validation failed", errs)
		return
	}

	ctx := r.Context()
	res, err := h.register.Exec(ctx, &usecase.RegisterUsecaseInput{
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		CountryCode: req.CountryCode,
		PhoneNumber: format.NormalizePhone(req.CountryCode, req.PhoneNumber),
	})
	if err != nil {
		h.log.Error(
			"register failed",
			"error", err,
			"payload", req,
		)
		h.responder.Error(w, exception.GetHttpCode(err), err.Error(), nil)
		return
	}

	h.responder.Success(w, http.StatusOK, res, "register successful", nil)
}

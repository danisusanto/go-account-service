package rest

import (
	"fmt"
	"net/http"

	"github.com/dani-susanto/go-account-service/internal/provider"
	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/json"
	"github.com/dani-susanto/go-common/jwt"
	"github.com/dani-susanto/go-common/log"
	commonrest "github.com/dani-susanto/go-common/rest"
	"github.com/dani-susanto/go-common/validator"
)

type handler struct {
	basePath   string
	env        env.Env
	router     commonrest.Rest
	validation validator.Validator
	json       json.JSON
	responder  commonrest.Responder
	log        log.Log
}

func (h *handler) extractUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, ok := jwt.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get user id from claims")
		h.responder.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return 0, false
	}
	return userID, true
}

func (h *handler) buildPath(path string) string {
	return fmt.Sprintf("%s%s", h.basePath, path)
}

func SetupHTTPRoutes(
	basePath string,
	env env.Env,
	usecase *provider.Usecases,
	router commonrest.Rest,
	validation validator.Validator,
	json json.JSON,
	responder commonrest.Responder,
	log log.Log,
) {
	handler := handler{
		basePath:   basePath,
		env:        env,
		router:     router,
		validation: validation,
		json:       json,
		responder:  responder,
		log:        log,
	}

	NewHealthHandler(handler).Register()
	NewSwaggerHandler(handler).Register()
	NewLoginHandler(handler, usecase.Login).Register()
	NewRegisterHandler(handler, usecase.Register).Register()
	NewEmailVerificationHandler(handler, usecase.EmailVerification).Register()
	NewMeHandler(handler, usecase.GetUser).Register()
	NewRefreshTokenHandler(handler, usecase.RefreshToken).Register()
	NewChangePasswordHandler(handler, usecase.ChangePassword).Register()
	NewEmailVerificationRequestHandler(handler, usecase.EmailVerificationRequest).Register()
	NewResetPasswordRequestHandler(handler, usecase.ResetPasswordRequest).Register()
	NewResetPasswordVerifyHandler(handler, usecase.ResetPasswordVerify).Register()
	NewResetPasswordHandler(handler, usecase.ResetPassword).Register()
	NewUpdateMeHandler(handler, usecase.UpdateUser).Register()
	NewChangeEmailRequestHandler(handler, usecase.ChangeEmailRequest).Register()
	NewChangeEmailVerifyHandler(handler, usecase.ChangeEmailVerify).Register()
}

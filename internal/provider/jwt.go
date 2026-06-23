package provider

import (
	"time"

	"github.com/dani-susanto/go-common/env"
	"github.com/dani-susanto/go-common/jwt"
)

func SetupAccessTokenJWT(env env.Env) (jwt.JWT, error) {
	accessTokenExpired, err := time.ParseDuration(env.Get("JWT_ACCESS_TOKEN_EXPIRED"))
	if err != nil {
		return nil, err
	}
	jwt := jwt.New(env.Get("JWT_ACCESS_TOKEN_SECRET_KEY"), accessTokenExpired)
	return jwt, nil
}

func SetupRefreshTokenJWT(env env.Env) (jwt.JWT, error) {
	refreshTokenExpired, err := time.ParseDuration(env.Get("JWT_REFRRESH_TOKEN_EXPIRED"))
	if err != nil {
		return nil, err
	}
	jwt := jwt.New(env.Get("JWT_REFRESH_TOKEN_SECRET_KEY"), refreshTokenExpired)
	return jwt, nil
}

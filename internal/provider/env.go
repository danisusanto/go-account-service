package provider

import (
	"os"

	"github.com/dani-susanto/go-common/env"
)

func SetupEnv() (env.Env, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	var envFile string

	switch appEnv {
	case "production":
		envFile = ".env.production"
	case "development":
		envFile = ".env.development"
	default:
		envFile = ".env.local"
	}

	env := env.New(envFile)
	if err := env.Load(); err != nil {
		return nil, err
	}

	return env, nil
}

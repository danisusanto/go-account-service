package main

import (
	"github.com/dani-susanto/go-account-service/internal/provider"
	"github.com/dani-susanto/go-account-service/seeders"
	"github.com/dani-susanto/go-common/log"
	"github.com/dani-susanto/go-common/postgres"
)

func main() {
	log := log.New()

	env, err := provider.SetupEnv()
	if err != nil {
		log.Error("failed to load env", "error", err.Error())
	}

	db, err := postgres.New(
		env.Get("DB_USERNAME"),
		env.Get("DB_PASSWORD"),
		env.Get("DB_HOST"),
		env.Get("DB_PORT"),
		env.Get("DB_DATABASE"),
	)
	if err != nil {
		log.Error("failed to load database", "error", err.Error())
	}

	seeders.RunSeeders(db)
}

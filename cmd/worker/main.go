package main

import (
	"log"

	"github.com/dani-susanto/go-account-service/internal/bootstrap"
)

func main() {
	server, err := bootstrap.NewEventServer()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

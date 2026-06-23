package main

import (
	"errors"

	"github.com/dani-susanto/go-common/log"
)

func main() {
	log := log.New()

	err := errors.New("internal server error")
	log.Error("failed to load database", "error", err.Error())
}

package main

import (
	"errors"

	"github.com/danisusanto/go-common/log"
)

func main() {
	log := log.New()

	err := errors.New("internal server error")
	log.Error("failed to load database", "error", err.Error())
}

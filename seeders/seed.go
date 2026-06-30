package seeders

import (
	"database/sql"
	"log"
	"reflect"
	"time"

	"github.com/danisusanto/go-common/hash"
)

type Seeder interface {
	Seed(db *sql.DB) error
}

func RunSeeders(db *sql.DB) {
	hash := hash.New()

	seeders := []Seeder{
		NewUserSeeder(hash),
	}

	for _, s := range seeders {
		start := time.Now()

		name := reflect.TypeOf(s).Name()

		if err := s.Seed(db); err != nil {
			log.Fatalf("[ERROR] %s error=%s", name, err.Error())
		}

		log.Printf("[INFO] %s took=%s", name, time.Since(start))
	}
}

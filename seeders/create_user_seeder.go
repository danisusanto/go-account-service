package seeders

import (
	"database/sql"

	"github.com/dani-susanto/go-common/hash"
)

type UserSeeder struct {
	hash hash.Hash
}

func (s *UserSeeder) Seed(db *sql.DB) error {
	hashedPassword, _ := s.hash.Hash("secret")

	_, err := db.Exec(
		`INSERT INTO users (name, email, password, country_code, phone_number, role, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW()) ON CONFLICT (email) DO NOTHING`,
		"John Doe",
		"john@email.com",
		hashedPassword,
		"62",
		"6281xxxx",
		"ADMIN",
	)
	return err
}

func NewUserSeeder(hash hash.Hash) *UserSeeder {
	return &UserSeeder{hash: hash}
}

package entity

import "time"

type Role string

const (
	ADMIN   Role = "ADMIN"
	MANAGER Role = "MANAGER"
	STAFF   Role = "STAFF"
	VIEWER  Role = "VIEWER"
)

type User struct {
	ID              int
	Name            string
	Email           string
	Password        string
	CountryCode     string
	PhoneNumber     string
	GoogleID        string
	IsActive        *bool
	Role            Role
	CreatedBy       int
	UpdatedBy       int
	DeletedBy       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EmailVerifiedAt *time.Time
	DeletedAt       time.Time
}

func NewUser(name, email, hashedPassword, countryCode, phone string) *User {
	isActive := true

	return &User{
		Name:        name,
		Email:       email,
		Password:    hashedPassword,
		CountryCode: countryCode,
		PhoneNumber: phone,
		Role:        ADMIN,
		IsActive:    &isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (e *User) SetEmailVerified() {
	now := time.Now()
	e.EmailVerifiedAt = &now
}

func (e *User) IsEmailVerified() bool {
	return e.EmailVerifiedAt != nil
}

func (u *User) IsUserActive() bool {
	return u.IsActive != nil && *u.IsActive
}

func (u *User) SetActive() {
	active := true
	u.IsActive = &active
}

func (u *User) SetInactive() {
	active := false
	u.IsActive = &active
}

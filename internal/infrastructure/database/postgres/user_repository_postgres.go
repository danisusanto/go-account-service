package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-common/log"
	commonpostgres "github.com/dani-susanto/go-common/postgres"
)

type userRepositoryPostgres struct {
	db  *sql.DB
	log log.Log
}

func NewUserRepositoryPostgres(
	db *sql.DB,
	log log.Log,
) repository.UserRepository {
	return &userRepositoryPostgres{
		db:  db,
		log: log,
	}
}

func (p *userRepositoryPostgres) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (
			name,
			email,
			password,
			country_code,
			phone_number,
			google_id,
			is_active,
			role,
			created_by,
			updated_by,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`

	err := p.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Password,
		user.CountryCode,
		user.PhoneNumber,
		user.GoogleID,
		user.IsActive,
		user.Role,
		user.CreatedBy,
		user.UpdatedBy,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *userRepositoryPostgres) Get(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT
			id,
			name,
			email,
			password,
			country_code,
			phone_number,
			is_active,
			role,
			created_at,
			updated_at,
			email_verified_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user entity.User

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CountryCode,
		&user.PhoneNumber,
		&user.IsActive,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerifiedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (p *userRepositoryPostgres) List(ctx context.Context) ([]entity.User, error) {
	panic("unimplemented")
}

func (p *userRepositoryPostgres) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT
			id,
			name,
			email,
			password,
			country_code,
			phone_number,
			is_active,
			role,
			created_at,
			updated_at,
			email_verified_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user entity.User

	err := p.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CountryCode,
		&user.PhoneNumber,
		&user.IsActive,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerifiedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (p *userRepositoryPostgres) Update(ctx context.Context, id int, user *entity.User) (*entity.User, error) {
	args := []any{}
	query := fmt.Sprintf(`
		UPDATE users
		%s
		%s
		RETURNING
			id,
			name,
			email,
			country_code,
			phone_number,
			is_active,
			role,
			created_at,
			updated_at,
			email_verified_at
		`,
		commonpostgres.BuildQueryUpdate(
			[]commonpostgres.QueryUpdate{
				{Field: "name", Value: user.Name},
				{Field: "country_code", Value: user.CountryCode},
				{Field: "phone_number", Value: user.PhoneNumber},
				{Field: "is_active", Value: user.IsActive},
				{Field: "role", Value: string(user.Role)},
				{Field: "updated_by", Value: user.UpdatedBy},
				{Field: "updated_at", Value: time.Now()},
			},
			&args,
		),
		commonpostgres.BuildQueryWhere(
			[]commonpostgres.QueryWhere{
				{Field: "id", Operator: commonpostgres.QueryOpEqual, Value: id},
				{Field: "deleted_at", Operator: commonpostgres.QueryOpIsNull},
			},
			&args,
		),
	)

	if err := p.db.QueryRowContext(
		context.Background(),
		query,
		args...,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CountryCode,
		&user.PhoneNumber,
		&user.IsActive,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.EmailVerifiedAt,
	); err != nil {
		return nil, err
	}

	return user, nil

}

func (p *userRepositoryPostgres) MarkEmailAsVerified(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET
			updated_by = $1,
			updated_at = NOW(),
			email_verified_at = $2
		WHERE id = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.EmailVerifiedAt,
		user.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *userRepositoryPostgres) SetPassword(ctx context.Context, id int, password string) error {
	query := `
		UPDATE users
		SET
			updated_by = $1,
			updated_at = NOW(),
			password = $2
		WHERE id = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		id,
		password,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *userRepositoryPostgres) SetEmail(ctx context.Context, id int, email string) error {
	query := `
		UPDATE users
		SET
			updated_by = $1,
			updated_at = NOW(),
			email = $2
		WHERE id = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		id,
		email,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *userRepositoryPostgres) Delete(ctx context.Context, id int) error {
	panic("unimplemented")
}

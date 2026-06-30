package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-common/log"
)

type userVerificationRepositoryPostgres struct {
	db  *sql.DB
	log log.Log
}

func NewUserVerificationRepositoryPostgres(
	db *sql.DB,
	log log.Log,
) repository.UserVerificationRepository {
	return &userVerificationRepositoryPostgres{
		db:  db,
		log: log,
	}
}

func (p *userVerificationRepositoryPostgres) Create(
	ctx context.Context,
	uev *entity.UserVerification,
) (*entity.UserVerification, error) {
	query := `
		INSERT INTO user_verifications (
			user_id,
			verification_type,
			token_hash,
			attempt,
			payload,
			created_by,
			updated_by,
			created_at,
			updated_at,
			expired_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`

	err := p.db.QueryRowContext(
		ctx,
		query,
		uev.UserID,
		uev.VerificationType,
		uev.TokenHash,
		uev.Attempt,
		uev.Payload,
		uev.UserID,
		uev.UserID,
		uev.CreatedAt,
		uev.UpdatedAt,
		uev.ExpiredAt,
	).Scan(&uev.ID)

	if err != nil {
		return nil, err
	}

	return uev, nil
}

func (p *userVerificationRepositoryPostgres) GetLatestActiveByUserID(
	ctx context.Context, userID int,
	verificationType entity.VerificationType,
) (*entity.UserVerification, error) {
	query := `
		SELECT
			id,
			user_id,
			verification_type,
			token_hash,
			attempt,
			payload
		FROM user_verifications
		WHERE
			deleted_at IS NULL
			AND expired_at > NOW()
			AND used_at IS NULL
			AND user_id = $1
			AND verification_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var uev entity.UserVerification

	err := p.db.QueryRowContext(
		ctx,
		query,
		userID,
		verificationType,
	).Scan(
		&uev.ID,
		&uev.UserID,
		&uev.VerificationType,
		&uev.TokenHash,
		&uev.Attempt,
		&uev.Payload,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &uev, nil
}

func (p *userVerificationRepositoryPostgres) GetLatestActiveVerifiedByUserID(
	ctx context.Context,
	userID int,
	verificationType entity.VerificationType,
) (*entity.UserVerification, error) {
	query := `
		SELECT
			id,
			user_id,
			verification_type,
			token_hash,
			attempt,
			payload
		FROM user_verifications
		WHERE
			deleted_at IS NULL
			AND expired_at > NOW()
			AND verified_at IS NOT NULL
			AND used_at IS NULL
			AND user_id = $1
			AND verification_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var uev entity.UserVerification

	err := p.db.QueryRowContext(
		ctx,
		query,
		userID,
		verificationType,
	).Scan(
		&uev.ID,
		&uev.UserID,
		&uev.VerificationType,
		&uev.TokenHash,
		&uev.Attempt,
		&uev.Payload,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &uev, nil
}

func (p *userVerificationRepositoryPostgres) IncrementAttempt(
	ctx context.Context,
	id int,
	userID int,
	verificationType entity.VerificationType,
) error {
	query := `
		UPDATE user_verifications
		SET
			attempt = attempt + 1,
			updated_by = $1,
			updated_at = NOW()
		WHERE id = $2
			AND verification_type = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		userID,
		id,
		verificationType,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *userVerificationRepositoryPostgres) MarkAsVerified(
	ctx context.Context,
	uev *entity.UserVerification,
) error {
	query := `
		UPDATE user_verifications
		SET
			updated_by = $1,
			updated_at = NOW(),
			verified_at = NOW()
		WHERE id = $2
			AND verification_type = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		uev.UserID,
		uev.ID,
		uev.VerificationType,
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *userVerificationRepositoryPostgres) MarkAsUsed(
	ctx context.Context,
	uev *entity.UserVerification,
) error {
	query := `
		UPDATE user_verifications
		SET
			updated_by = $1,
			updated_at = NOW(),
			used_at = NOW()
		WHERE id = $2
			AND verification_type = $3
	`
	_, err := p.db.ExecContext(
		ctx,
		query,
		uev.UserID,
		uev.ID,
		uev.VerificationType,
	)
	if err != nil {
		return err
	}
	return nil
}

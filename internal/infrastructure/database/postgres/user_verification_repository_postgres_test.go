package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-account-service/internal/infrastructure/database/postgres"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	"github.com/stretchr/testify/assert"
)

type mockTestUserVerificationRepository struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
	log  *mocklog.MockLog
}

func newMockTestUserVerificationRepository(t *testing.T) mockTestUserVerificationRepository {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return mockTestUserVerificationRepository{
		db:   db,
		mock: mock,
		log:  mocklog.NewMockLog(t),
	}
}

func TestUserVerificationRepositoryPostgres_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	expiredAt := now.Add(2 * time.Minute)

	uev := &entity.UserVerification{
		UserID:           1,
		VerificationType: entity.ChangeEmail,
		TokenHash:        "hashed_token",
		Attempt:          0,
		Payload:          "test@example.com",
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiredAt:        expiredAt,
	}

	tests := []struct {
		name      string
		input     *entity.UserVerification
		output    *entity.UserVerification
		err       error
		setupMock func(m mockTestUserVerificationRepository)
	}{
		{
			name:  "failed to create user verification",
			input: uev,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`INSERT INTO user_verifications`).
					WithArgs(
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
					).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			input: uev,
			output: func() *entity.UserVerification {
				u := *uev
				u.ID = 1
				return &u
			}(),
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`INSERT INTO user_verifications`).
					WithArgs(
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
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			res, err := repo.Create(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserVerificationRepositoryPostgres_GetLatestActiveByUserID(t *testing.T) {
	ctx := context.Background()

	uev := &entity.UserVerification{
		ID:               1,
		UserID:           1,
		VerificationType: entity.ChangeEmail,
		TokenHash:        "hashed_token",
		Attempt:          0,
		Payload:          "test@example.com",
	}

	tests := []struct {
		name             string
		userID           int
		verificationType entity.VerificationType
		output           *entity.UserVerification
		err              error
		setupMock        func(m mockTestUserVerificationRepository)
	}{
		{
			name:             "not found",
			userID:           1,
			verificationType: entity.ChangeEmail,
			err:              repository.ErrNotFound,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:             "db error",
			userID:           1,
			verificationType: entity.ChangeEmail,
			err:              sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:             "success",
			userID:           1,
			verificationType: entity.ChangeEmail,
			output:           uev,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "user_id", "verification_type", "token_hash", "attempt", "payload",
					}).AddRow(
						uev.ID, uev.UserID, uev.VerificationType, uev.TokenHash, uev.Attempt, uev.Payload,
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			res, err := repo.GetLatestActiveByUserID(ctx, tt.userID, tt.verificationType)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserVerificationRepositoryPostgres_GetLatestActiveVerifiedByUserID(t *testing.T) {
	ctx := context.Background()

	uev := &entity.UserVerification{
		ID:               1,
		UserID:           1,
		VerificationType: entity.ChangeEmail,
		TokenHash:        "hashed_token",
		Attempt:          0,
		Payload:          "test@example.com",
	}

	tests := []struct {
		name             string
		userID           int
		verificationType entity.VerificationType
		output           *entity.UserVerification
		err              error
		setupMock        func(m mockTestUserVerificationRepository)
	}{
		{
			name:             "not found",
			userID:           1,
			verificationType: entity.ChangeEmail,
			err:              repository.ErrNotFound,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:             "db error",
			userID:           1,
			verificationType: entity.ChangeEmail,
			err:              sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:             "success",
			userID:           1,
			verificationType: entity.ChangeEmail,
			output:           uev,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1, entity.ChangeEmail).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "user_id", "verification_type", "token_hash", "attempt", "payload",
					}).AddRow(
						uev.ID, uev.UserID, uev.VerificationType, uev.TokenHash, uev.Attempt, uev.Payload,
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			res, err := repo.GetLatestActiveVerifiedByUserID(ctx, tt.userID, tt.verificationType)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserVerificationRepositoryPostgres_IncrementAttempt(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		id               int
		userID           int
		verificationType entity.VerificationType
		err              error
		setupMock        func(m mockTestUserVerificationRepository)
	}{
		{
			name:             "db error",
			id:               1,
			userID:           1,
			verificationType: entity.ChangeEmail,
			err:              sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(1, 1, entity.ChangeEmail).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:             "success",
			id:               1,
			userID:           1,
			verificationType: entity.ChangeEmail,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(1, 1, entity.ChangeEmail).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			err := repo.IncrementAttempt(ctx, tt.id, tt.userID, tt.verificationType)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserVerificationRepositoryPostgres_MarkAsVerified(t *testing.T) {
	ctx := context.Background()

	uev := &entity.UserVerification{
		ID:               1,
		UserID:           1,
		VerificationType: entity.ChangeEmail,
	}

	tests := []struct {
		name      string
		input     *entity.UserVerification
		err       error
		setupMock func(m mockTestUserVerificationRepository)
	}{
		{
			name:  "db error",
			input: uev,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(uev.UserID, uev.ID, uev.VerificationType).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			input: uev,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(uev.UserID, uev.ID, uev.VerificationType).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			err := repo.MarkAsVerified(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserVerificationRepositoryPostgres_MarkAsUsed(t *testing.T) {
	ctx := context.Background()

	uev := &entity.UserVerification{
		ID:               1,
		UserID:           1,
		VerificationType: entity.ChangeEmail,
	}

	tests := []struct {
		name      string
		input     *entity.UserVerification
		err       error
		setupMock func(m mockTestUserVerificationRepository)
	}{
		{
			name:  "db error",
			input: uev,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(uev.UserID, uev.ID, uev.VerificationType).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			input: uev,
			setupMock: func(m mockTestUserVerificationRepository) {
				m.mock.ExpectExec(`UPDATE user_verifications`).
					WithArgs(uev.UserID, uev.ID, uev.VerificationType).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserVerificationRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserVerificationRepositoryPostgres(m.db, m.log)
			err := repo.MarkAsUsed(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

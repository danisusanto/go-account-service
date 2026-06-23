package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dani-susanto/go-account-service/internal/domain/entity"
	"github.com/dani-susanto/go-account-service/internal/domain/repository"
	"github.com/dani-susanto/go-account-service/internal/infrastructure/database/postgres"
	mocklog "github.com/dani-susanto/go-account-service/mocks/log"
	"github.com/stretchr/testify/assert"
)

type mockTestUserRepository struct {
	db   *sql.DB
	mock sqlmock.Sqlmock
	log  *mocklog.MockLog
}

func newMockTestUserRepository(t *testing.T) mockTestUserRepository {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return mockTestUserRepository{
		db:   db,
		mock: mock,
		log:  mocklog.NewMockLog(t),
	}
}

func TestUserRepositoryPostgres_Create(t *testing.T) {
	ctx := context.Background()
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "6212345678")

	tests := []struct {
		name      string
		input     *entity.User
		output    *entity.User
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "failed to create user",
			input: user,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`INSERT INTO users`).
					WithArgs(
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
					).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			input: user,
			output: func() *entity.User {
				u := *user
				u.ID = 1
				return &u
			}(),
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`INSERT INTO users`).
					WithArgs(
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
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			res, err := repo.Create(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_Get(t *testing.T) {
	ctx := context.Background()
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "6212345678")

	tests := []struct {
		name      string
		input     int
		output    *entity.User
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "not found",
			input: 1,
			err:   repository.ErrNotFound,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:  "db error",
			input: 1,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:   "success",
			input:  1,
			output: user,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "name", "email", "password", "country_code", "phone_number",
						"is_active", "role", "created_at", "updated_at", "email_verified_at",
					}).AddRow(
						user.ID, user.Name, user.Email, user.Password, user.CountryCode, user.PhoneNumber,
						user.IsActive, user.Role, user.CreatedAt, user.UpdatedAt, user.EmailVerifiedAt,
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			res, err := repo.Get(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_FindByEmail(t *testing.T) {
	ctx := context.Background()
	user := entity.NewUser("test", "test@example.com", "hashedpassword", "62", "6212345678")

	tests := []struct {
		name      string
		input     string
		output    *entity.User
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "not found",
			input: "notfound@example.com",
			err:   repository.ErrNotFound,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs("notfound@example.com").
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:  "db error",
			input: "test@example.com",
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs("test@example.com").
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:   "success",
			input:  "test@example.com",
			output: user,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`SELECT`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "name", "email", "password", "country_code", "phone_number",
						"is_active", "role", "created_at", "updated_at", "email_verified_at",
					}).AddRow(
						user.ID, user.Name, user.Email, user.Password, user.CountryCode, user.PhoneNumber,
						user.IsActive, user.Role, user.CreatedAt, user.UpdatedAt, user.EmailVerifiedAt,
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			res, err := repo.FindByEmail(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_Update(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	user := entity.NewUser("test updated", "test@example.com", "hashedpassword", "62", "6212345678")
	user.ID = 1

	updatedUser := &entity.User{
		ID:          1,
		Name:        "test updated",
		Email:       "test@example.com",
		Password:    "hashedpassword",
		CountryCode: "62",
		PhoneNumber: "6212345678",
		IsActive:    user.IsActive,
		Role:        user.Role,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tests := []struct {
		name      string
		id        int
		input     *entity.User
		output    *entity.User
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "db error",
			id:    1,
			input: user,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`UPDATE users`).
					WithArgs(
						user.Name,
						user.CountryCode,
						user.PhoneNumber,
						user.IsActive,
						string(user.Role),
						sqlmock.AnyArg(), // updated_at (time.Now())
						1,                // id dari WHERE
					).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:   "success",
			id:     1,
			input:  user,
			output: updatedUser,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectQuery(`UPDATE users`).
					WithArgs(
						user.Name,
						user.CountryCode,
						user.PhoneNumber,
						user.IsActive,
						string(user.Role),
						sqlmock.AnyArg(), // updated_at (time.Now())
						1,                // id dari WHERE
					).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "name", "email", "country_code", "phone_number",
						"is_active", "role", "created_at", "updated_at", "email_verified_at",
					}).AddRow(
						updatedUser.ID, updatedUser.Name, updatedUser.Email,
						updatedUser.CountryCode, updatedUser.PhoneNumber,
						updatedUser.IsActive, updatedUser.Role,
						updatedUser.CreatedAt, updatedUser.UpdatedAt, updatedUser.EmailVerifiedAt,
					))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			res, err := repo.Update(ctx, tt.id, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_MarkEmailAsVerified(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	user := &entity.User{
		ID:              1,
		EmailVerifiedAt: &now,
	}

	tests := []struct {
		name      string
		input     *entity.User
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "db error",
			input: user,
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(user.ID, user.EmailVerifiedAt, user.ID).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			input: user,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(user.ID, user.EmailVerifiedAt, user.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			err := repo.MarkEmailAsVerified(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_SetPassword(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		id        int
		password  string
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:     "db error",
			id:       1,
			password: "newhashedpassword",
			err:      sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(1, "newhashedpassword", 1).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:     "success",
			id:       1,
			password: "newhashedpassword",
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(1, "newhashedpassword", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			err := repo.SetPassword(ctx, tt.id, tt.password)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepositoryPostgres_SetEmail(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		id        int
		email     string
		err       error
		setupMock func(m mockTestUserRepository)
	}{
		{
			name:  "db error",
			id:    1,
			email: "newemail@example.com",
			err:   sql.ErrConnDone,
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(1, "newemail@example.com", 1).
					WillReturnError(sql.ErrConnDone)
			},
		},
		{
			name:  "success",
			id:    1,
			email: "newemail@example.com",
			setupMock: func(m mockTestUserRepository) {
				m.mock.ExpectExec(`UPDATE users`).
					WithArgs(1, "newemail@example.com", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockTestUserRepository(t)
			defer m.db.Close()

			if tt.setupMock != nil {
				tt.setupMock(m)
			}

			repo := postgres.NewUserRepositoryPostgres(m.db, m.log)
			err := repo.SetEmail(ctx, tt.id, tt.email)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, m.mock.ExpectationsWereMet())
		})
	}
}

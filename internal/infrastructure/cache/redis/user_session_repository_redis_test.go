package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/danisusanto/go-account-service/internal/domain/entity"
	"github.com/danisusanto/go-account-service/internal/domain/repository"
	"github.com/danisusanto/go-account-service/internal/infrastructure/cache/redis"
	mocklog "github.com/danisusanto/go-account-service/mocks/log"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

type mockTestUserSessionRepository struct {
	mock redismock.ClientMock
	log  *mocklog.MockLog
}

func TestUserSessionRepository_Create(t *testing.T) {
	ctx := context.Background()
	expiredAt := time.Now().Add(time.Hour)

	session := &entity.UserSession{
		UserID:      1,
		SessionID:   "session-id",
		SessionType: entity.AccessToken,
		ExpiredAt:   &expiredAt,
	}
	key := entity.NewSessionKey(session.SessionType, session.SessionID)

	tests := []struct {
		name      string
		input     *entity.UserSession
		output    *entity.UserSession
		err       error
		setupMock func(m mockTestUserSessionRepository)
	}{
		{
			name:  "failed to hset",
			input: session,
			err:   assert.AnError,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHSet(key, session).SetErr(assert.AnError)
			},
		},
		{
			name:  "failed to set expiry",
			input: session,
			err:   assert.AnError,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHSet(key, session).SetVal(1)
				m.mock.ExpectExpire(key, time.Until(expiredAt)).SetErr(assert.AnError)
			},
		},
		{
			name:   "success",
			input:  session,
			output: session,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHSet(key, session).SetVal(1)
				m.mock.ExpectExpire(key, time.Until(expiredAt)).SetVal(true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			log := mocklog.NewMockLog(t)

			if tt.setupMock != nil {
				tt.setupMock(mockTestUserSessionRepository{mock: mock, log: log})
			}

			repo := redis.NewUserSessionRepository(client, log)
			res, err := repo.Create(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserSessionRepository_CreateBulk(t *testing.T) {
	ctx := context.Background()
	expiredAt := time.Now().Add(time.Hour)

	sessions := []*entity.UserSession{
		{
			UserID:      1,
			SessionID:   "access-session-id",
			SessionType: entity.AccessToken,
			ExpiredAt:   &expiredAt,
		},
		{
			UserID:      1,
			SessionID:   "refresh-session-id",
			SessionType: entity.RefreshToken,
			ExpiredAt:   &expiredAt,
		},
	}

	tests := []struct {
		name      string
		input     []*entity.UserSession
		output    []*entity.UserSession
		err       error
		setupMock func(m mockTestUserSessionRepository)
	}{
		{
			name:   "success",
			input:  sessions,
			output: sessions,
			setupMock: func(m mockTestUserSessionRepository) {
				for _, s := range sessions {
					key := entity.NewSessionKey(s.SessionType, s.SessionID)
					m.mock.ExpectHSet(key, s).SetVal(1)
					m.mock.ExpectExpire(key, time.Until(expiredAt)).SetVal(true)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			log := mocklog.NewMockLog(t)

			if tt.setupMock != nil {
				tt.setupMock(mockTestUserSessionRepository{mock: mock, log: log})
			}

			repo := redis.NewUserSessionRepository(client, log)
			res, err := repo.CreateBulk(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserSessionRepository_FindByKey(t *testing.T) {
	ctx := context.Background()

	session := &entity.UserSession{
		UserID:      1,
		SessionID:   "session-id",
		SessionType: entity.AccessToken,
	}
	key := entity.NewSessionKey(session.SessionType, session.SessionID)

	tests := []struct {
		name      string
		input     string
		output    *entity.UserSession
		err       error
		setupMock func(m mockTestUserSessionRepository)
	}{
		{
			name:  "db error",
			input: key,
			err:   assert.AnError,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHGetAll(key).SetErr(assert.AnError)
			},
		},
		{
			name:  "not found",
			input: key,
			err:   repository.ErrNotFound,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHGetAll(key).SetVal(map[string]string{})
			},
		},
		{
			name:   "success",
			input:  key,
			output: session,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectHGetAll(key).SetVal(map[string]string{
					"user_id":      "1",
					"session_id":   "session-id",
					"session_type": string(entity.AccessToken),
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			log := mocklog.NewMockLog(t)

			if tt.setupMock != nil {
				tt.setupMock(mockTestUserSessionRepository{mock: mock, log: log})
			}

			repo := redis.NewUserSessionRepository(client, log)
			res, err := repo.FindByKey(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.output, res)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserSessionRepository_DeleteByKey(t *testing.T) {
	ctx := context.Background()
	key := "access_token:session-id"

	tests := []struct {
		name      string
		input     string
		err       error
		setupMock func(m mockTestUserSessionRepository)
	}{
		{
			name:  "db error",
			input: key,
			err:   assert.AnError,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectDel(key).SetErr(assert.AnError)
			},
		},
		{
			name:  "success",
			input: key,
			setupMock: func(m mockTestUserSessionRepository) {
				m.mock.ExpectDel(key).SetVal(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			log := mocklog.NewMockLog(t)

			if tt.setupMock != nil {
				tt.setupMock(mockTestUserSessionRepository{mock: mock, log: log})
			}

			repo := redis.NewUserSessionRepository(client, log)
			err := repo.DeleteByKey(ctx, tt.input)

			assert.Equal(t, tt.err, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

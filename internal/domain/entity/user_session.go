package entity

import (
	"fmt"
	"strings"
	"time"
)

const (
	AccessToken  = "ACCESS_TOKEN"
	RefreshToken = "REFRESH_TOKEN"
	SessionKey   = "account:user:session"
)

type UserSession struct {
	ID          int        `redis:"-"`
	UserID      int        `redis:"user_id"`
	SessionID   string     `redis:"session_id"`
	SessionType string     `redis:"session_type"`
	CreatedBy   int        `redis:"created_by"`
	UpdatedBy   int        `redis:"updated_by"`
	DeletedBy   int        `redis:"-"`
	CreatedAt   time.Time  `redis:"created_at"`
	UpdatedAt   time.Time  `redis:"updated_at"`
	ExpiredAt   *time.Time `redis:"expired_at"`
	DeletedAt   *time.Time `redis:"-"`
}

func NewSession(
	userID int,
	sessionID string,
	sessionType string,
	expiredAt *time.Time,
) *UserSession {
	return &UserSession{
		UserID:      userID,
		SessionID:   sessionID,
		SessionType: sessionType,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiredAt:   expiredAt,
	}
}

func NewSessionKey(
	sessionType string,
	sessionID string,
) string {
	switch sessionType {
	case RefreshToken:
		return RefreshTokenKey(sessionID)
	default:
		return AccessTokenKey(sessionID)
	}
}

func AccessTokenKey(sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", SessionKey, strings.ToLower(AccessToken), sessionID)
}

func RefreshTokenKey(sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", SessionKey, strings.ToLower(RefreshToken), sessionID)
}

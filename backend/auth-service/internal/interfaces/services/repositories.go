package services

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
)

type SessionRepo interface {
	SaveSession(ctx context.Context, sessionID string, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type AuthRepo interface {
	SaveUser(ctx context.Context, user *models.User) (string, error)
	GetUserByUserName(ctx context.Context, username string) (*models.User, error)
	CheckUserExists(ctx context.Context, username string) (bool, error)
}

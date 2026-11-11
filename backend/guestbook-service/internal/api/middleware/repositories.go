package middleware

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
)

type SessionRepo interface {
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

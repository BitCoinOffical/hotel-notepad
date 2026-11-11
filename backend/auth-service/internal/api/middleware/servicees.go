package middleware

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterDTO) (string, error)
	Login(ctx context.Context, req *dto.LoginDTO) (string, error)
	Logout(ctx context.Context, sessionID string) error

	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

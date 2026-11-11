package handlers

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/services"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterDTO) (string, error)
	Login(ctx context.Context, req *dto.LoginDTO) (string, error)
	Logout(ctx context.Context, sessionID string) error

	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
}

type Services struct {
	Auth *services.AuthService
}

func NewServices(pool *pgxpool.Pool, rdb *redis.Client) *Services {
	session := sessions.NewSessionRepo(rdb)
	authRepo := repo.NewAuthRepo(pool)

	auth := services.NewAuthService(authRepo, session)
	return &Services{Auth: auth}
}

type Handlers struct {
	Auth *AuthHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger) *Handlers {
	auth := NewAuthHandler(srvc.Auth, logger)
	return &Handlers{
		Auth: auth,
	}
}

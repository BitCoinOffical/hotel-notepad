package handlers

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/interfaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type GuestBookService interface {
	CreateEntry(ctx context.Context, req *dto.CreateGuestDTO) error
	GetEntries(ctx context.Context, userID string) ([]models.Guest, error)
	GetEntryByID(ctx context.Context, id, userID string) (*models.Guest, error)
	UpdateEntry(ctx context.Context, id, userID string, req *dto.UpdateGuestDTO) (*models.Guest, error)
	DeleteEntry(ctx context.Context, id, userID string) error
	SearchEntry(ctx context.Context, userID string, firstName, midName, lastName string) (*models.Guest, error)
}

type Services struct {
	Guest *services.GuestBookService
}

func NewServices(pool *pgxpool.Pool) *Services {
	repo := repo.NewGuestBookRepo(pool)

	guest := services.NewGuestBookHandler(repo)
	return &Services{Guest: guest}
}

type Handlers struct {
	Guest *GuestBookHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger) *Handlers {
	guest := NewGuestBookHandler(logger, srvc.Guest)
	return &Handlers{
		Guest: guest,
	}
}

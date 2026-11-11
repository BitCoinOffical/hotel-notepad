package services

import (
	"context"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
)

type GuestBookRepo interface {
	CreateEntry(ctx context.Context, guest *models.Guest) error
	GetEntries(ctx context.Context, userID string) ([]models.Guest, error)
	GetEntryByID(ctx context.Context, id, userID string) (*models.Guest, error)
	UpdateEntry(ctx context.Context, id, userID string, guest *models.UpdateGuest) (*models.Guest, error)
	DeleteEntry(ctx context.Context, id, userID string) error
	SearchEntry(ctx context.Context, userID string, firstName, midName, lastName string) (*models.Guest, error)
}

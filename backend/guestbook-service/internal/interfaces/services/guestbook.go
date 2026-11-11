package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
)

type GuestBookService struct {
	repo GuestBookRepo
}

func NewGuestBookHandler(repo GuestBookRepo) *GuestBookService {
	return &GuestBookService{repo: repo}
}

func (s *GuestBookService) CreateEntry(ctx context.Context, req *dto.CreateGuestDTO) error {
	guest := models.Guest{
		FirstName:     req.FirstName,
		MidName:       req.MidName,
		LastName:      req.LastName,
		Note:          *req.Note,
		Phone:         req.Phone,
		CurrentDate:   req.CurrentDate,
		DepartureDate: req.DepartureDate,
	}
	if err := s.repo.CreateEntry(ctx, &guest); err != nil {
		return fmt.Errorf("s.repo.CreateEntry: %w", err)
	}
	return nil
}

func (s *GuestBookService) GetEntries(ctx context.Context, userID string) ([]models.Guest, error) {
	guests, err := s.repo.GetEntries(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetEntries: %w", err)
	}
	return guests, nil
}
func (s *GuestBookService) GetEntryByID(ctx context.Context, id, userID string) (*models.Guest, error) {
	guest, err := s.repo.GetEntryByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetEntryByID:%w", err)
	}
	return guest, nil
}
func (s *GuestBookService) UpdateEntry(ctx context.Context, id, userID string, req *dto.UpdateGuestDTO) (*models.Guest, error) {
	guest := models.UpdateGuest{
		FirstName:     req.FirstName,
		MidName:       req.MidName,
		LastName:      req.LastName,
		Note:          req.Note,
		Phone:         req.Phone,
		CurrentDate:   req.CurrentDate,
		DepartureDate: req.DepartureDate,
	}
	newguest, err := s.repo.UpdateEntry(ctx, id, userID, &guest)
	if err != nil {
		return nil, fmt.Errorf("s.repo.UpdateEntry: %w", err)
	}
	return newguest, nil
}
func (s *GuestBookService) DeleteEntry(ctx context.Context, id, userID string) error {
	if err := s.repo.DeleteEntry(ctx, id, userID); err != nil {
		return fmt.Errorf("s.repo.DeleteEntry: %w", err)
	}
	return nil
}

func (s *GuestBookService) SearchEntry(ctx context.Context, userID string, firstName, midName, lastName string) (*models.Guest, error) {
	guest, err := s.repo.SearchEntry(ctx, userID, firstName, midName, lastName)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SearchEntry: %w", err)
	}
	return guest, nil
}

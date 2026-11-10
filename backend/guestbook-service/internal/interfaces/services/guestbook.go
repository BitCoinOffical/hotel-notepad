package services

import (
	"context"
)

type GuestBookService struct {
}

func NewGuestBookHandler() *GuestBookService {
	return &GuestBookService{}
}

func (s *GuestBookService) CreateEntry(ctx context.Context) {

}

func (s *GuestBookService) GetEntries(ctx context.Context) {

}
func (s *GuestBookService) GetEntryByID(ctx context.Context) {

}
func (s *GuestBookService) UpdateEntry(ctx context.Context) {

}
func (s *GuestBookService) DeleteEntry(ctx context.Context) {

}

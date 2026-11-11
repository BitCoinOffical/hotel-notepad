package dto

import "time"

type CreateGuestDTO struct {
	FirstName     string    `json:"first_name"  binding:"max=255,required"`
	MidName       string    `json:"mid_name"  binding:"max=255,required"`
	LastName      string    `json:"last_name"  binding:"max=255,required"`
	Note          *string   `json:"note" binding:"omitempty,max=2048"`
	Phone         int       `json:"phone" binding:"required"`
	ArrivalDate   time.Time `json:"arrival_date"  binding:"required"`
	DepartureDate time.Time `json:"departure_date"  binding:"required"`
}

type UpdateGuestDTO struct {
	FirstName     *string    `json:"first_name" binding:"omitempty,max=255"`
	MidName       *string    `json:"mid_name" binding:"omitempty,max=255"`
	LastName      *string    `json:"last_name" binding:"omitempty,max=255"`
	Note          *string    `json:"note" binding:"omitempty,max=2048"`
	Phone         *int       `json:"phone" binding:"omitempty"`
	ArrivalDate   *time.Time `json:"arrival_date" binding:"omitempty"`
	DepartureDate *time.Time `json:"departure_date" binding:"omitempty"`
}

package dto

import "time"

type CreateGuestDTO struct {
	FirstName     string    `json:"first_name"  binding:"max=255,required"`
	MidName       string    `json:"mid_name"  binding:"max=255,required"`
	LastName      string    `json:"last_name"  binding:"max=255,required"`
	Note          *string   `json:"note"  binding:"max=2048,omniempty"`
	Phone         int       `json:"phone"  binding:"len=11,required"`
	CurrentDate   time.Time `json:"current_date"  binding:"required"`
	DepartureDate time.Time `json:"departure_date"  binding:"required"`
}

type UpdateGuestDTO struct {
	FirstName     *string    `json:"first_name" binding:"max=255,omniempty"`
	MidName       *string    `json:"mid_name" binding:"max=255,omniempty"`
	LastName      *string    `json:"last_name" binding:"max=255,omniempty"`
	Note          *string    `json:"note" binding:"max=2048,omniempty"`
	Phone         *int       `json:"phone" binding:"len=11,omniempty"`
	CurrentDate   *time.Time `json:"current_date" binding:"omniempty"`
	DepartureDate *time.Time `json:"departure_date" binding:"omniempty"`
}

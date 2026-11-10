package models

import "time"

type Guest struct {
	GuestID       string    `db:"id"`
	UserID        string    `db:"user_id"`
	FirstName     string    `db:"first_name"`
	MidName       string    `db:"mid_name"`
	LastName      string    `db:"last_name"`
	Note          string    `db:"note"`
	Phone         int       `db:"phone"`
	IsDelete      bool      `db:"is_delete"`
	CurrentDate   time.Time `db:"current_date"`
	DepartureDate time.Time `db:"departure_date"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UpdateGuest struct {
	FirstName     *string    `db:"first_name"`
	MidName       *string    `db:"mid_name"`
	LastName      *string    `db:"last_name"`
	Note          *string    `db:"note"`
	Phone         *int       `db:"phone"`
	CurrentDate   *time.Time `db:"current_date"`
	DepartureDate *time.Time `db:"departure_date"`
}

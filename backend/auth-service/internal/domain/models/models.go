package models

import "time"

type User struct {
	ID           string    `db:"id"`
	UserName     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Session struct {
	UserID    string
	Username  string
	Role      string
	CreatedAt time.Time
}

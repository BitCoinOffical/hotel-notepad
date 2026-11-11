package models

import "time"

type Session struct {
	UserID    string
	Username  string
	Role      string
	CreatedAt time.Time
}

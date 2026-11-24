package entities

import "time"

type Question struct {
	ID        int
	UserID    int
	Question  string
	Answered  bool
	ReturnAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

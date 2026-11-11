package entities

import "time"

type Question struct {
	ID        int
	UserID    int
	Question  string
	Answered  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

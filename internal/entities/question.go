package entities

import "time"

type Question struct {
	ID        int
	Question  string
	Answered  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

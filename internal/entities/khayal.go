package entities

import "time"

type Khayal struct {
	ID         int
	QuestionID int
	Khayal     string
	CreatedAt  time.Time
}

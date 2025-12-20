package data

import (
	"time"
)

type Khayal struct {
	ID        int       `json:"id"`
	Question  Question  `json:"question"`
	Khayal    string    `json:"khayal"`
	CreatedAt time.Time `json:"created"`
}

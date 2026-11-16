package entities

import (
	"time"

	"github.com/google/uuid"
)

type Question struct {
	ID                 int
	UserID             int
	ClientID           uuid.UUID
	QuestionCiphertext []byte
	EncryptionVersion  int
	Answered           bool
	ReturnAt           time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type LegacyQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
}

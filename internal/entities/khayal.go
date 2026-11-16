package entities

import (
	"time"

	"github.com/google/uuid"
)

type Khayal struct {
	ID                int
	UserID            int
	QuestionID        int
	ClientID          uuid.UUID
	KhayalCiphertext  []byte
	EncryptionVersion int
	CreatedAt         time.Time
}

type LegacyKhayal struct {
	ID         int    `json:"id"`
	QuestionID int    `json:"question_id"`
	Khayal     string `json:"khayal"`
}

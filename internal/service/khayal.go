package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/google/uuid"
)

var ErrEmptyKhayal = errors.New("khayal cannot be empty")

type KhayalRepository interface {
	CreateKhayal(ctx context.Context, userID, questionID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int, returnAt time.Time) error
	GetAllKhayals(ctx context.Context, userID, questionID int) ([]entities.Khayal, error)
}

type KhayalService struct {
	KhayalRepo   KhayalRepository
	QuestionRepo QuestionRepository
}

func NewKhayalService(KhayalRepo KhayalRepository, QuestionRepo QuestionRepository) *KhayalService {
	return &KhayalService{
		KhayalRepo:   KhayalRepo,
		QuestionRepo: QuestionRepo,
	}
}

func (s *KhayalService) CreateKhayal(ctx context.Context, userID, questionID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error {
	if len(ciphertext) == 0 {
		return ErrEmptyKhayal
	}
	if len(ciphertext) < 12+16 || encryptionVersion < 1 {
		return ErrInvalidCiphertext
	}
	if err := s.KhayalRepo.CreateKhayal(ctx, userID, questionID, clientID, ciphertext, encryptionVersion, time.Now().Add(1*time.Minute)); err != nil {
		return fmt.Errorf("create khayal: %w", err)
	}
	return nil
}

func (s *KhayalService) GetAllKhayals(ctx context.Context, userID, questionID int) ([]entities.Khayal, error) {
	khayals, err := s.KhayalRepo.GetAllKhayals(ctx, userID, questionID)
	if err != nil {
		return nil, fmt.Errorf("get all khayals for question %d: %w", questionID, err)
	}
	return khayals, nil
}
